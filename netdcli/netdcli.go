/***
Copyright 2014 Cisco Systems Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
    "strings"
    "strconv"
    "net"

	"github.com/contiv/netplugin/core"
	"github.com/contiv/netplugin/drivers"
	"github.com/contiv/netplugin/netutils"
	"github.com/contiv/netplugin/gstate"
)

const (
	CLI_CONSTRUCT_GLOBAL = "global"
	CLI_CONSTRUCT_NW = "network"
	CLI_CONSTRUCT_EP = "endpoint"
	CLI_OPER_GET     = "get"
	CLI_OPER_CREATE  = "create"
	CLI_OPER_DELETE  = "delete"
	CLI_OPER_ATTACH  = "attach"
	CLI_OPER_DETACH  = "detach"
)

var validOperList = []string { CLI_OPER_GET, CLI_OPER_CREATE, CLI_OPER_DELETE, CLI_OPER_ATTACH, CLI_OPER_DETACH }

type CliError struct {
	Desc string
}

func (e *CliError) Error() string {
	return e.Desc
}

type Operation struct {
	val string
}

func (o *Operation) isValid(val string) bool {
    for _, str := range validOperList {
        if str == val {
            return true
        }
    }
    return false
}

func (o *Operation) String() string {
	return fmt.Sprintf("%s ", validOperList)
}

func (o *Operation) Set(val string) error {
	if ! o.isValid(val) {
		return &CliError{
            Desc: fmt.Sprintf("invalid value for construct (%s). Allowed values: %s",
			val, o.String())}
	}
	o.val = val
	return nil
}

func (o *Operation) Get() interface{} {
	return o.val
}

type Construct struct {
	val string
}

func (c *Construct) String() string {
	return fmt.Sprintf("%s or %s", CLI_CONSTRUCT_NW, CLI_CONSTRUCT_EP)
}

func (c *Construct) Set(val string) error {
	if val != CLI_CONSTRUCT_NW && val != CLI_CONSTRUCT_EP && val != CLI_CONSTRUCT_GLOBAL {
		return &CliError{Desc: fmt.Sprintf("invalid value for construct (%s). Allowed values: %s",
			val, c.String())}
	}
	c.val = val
	return nil
}

func (c *Construct) Get() interface{} {
	return c.val
}

type cliOpts struct {
    help            bool
    oper            Operation
    etcdUrl         string
    construct       Construct
    netId           string
    pktTag          string
    pktTagType      string
    subnetCidr      string
    ipAddr          string
    contName        string
    subnetIp        string
    subnetLen       uint
    allocSubnetLen  uint
    defaultGw       string
    idStr           string
    vlans           string
    vxlans          string
    homingHost      string
}

var opts cliOpts
var flagSet *flag.FlagSet

func init() {
	flagSet = flag.NewFlagSet("netdcli", flag.ExitOnError)
	flagSet.Var(&opts.oper,
        "oper",
        "Operation to perform")
	flagSet.Var(&opts.construct,
        "construct",
        "Construct to operate on i.e network or endpoint")
	flagSet.StringVar(&opts.etcdUrl,
        "etcd-url",
        "http://127.0.0.1:4001",
        "Etcd cluster url")
	flagSet.StringVar(&opts.netId,
        "net-id",
        "",
        "Network id of the endpoint")
	flagSet.StringVar(&opts.pktTag,
        "tag",
        "auto",
        "Vlan/Vxlan tag of the network")
	flagSet.StringVar(&opts.pktTagType,
        "tag-type",
        "vlan",
        "Vlan/Vxlan tag of the network")
	flagSet.StringVar(&opts.subnetCidr,
        "subnet",
        "",
        "Network Subnet IP with mask e.g. 11.0.1.1/24, or 0/24 to specify only mask")
	flagSet.StringVar(&opts.defaultGw,
        "gw",
        "",
        "Default Gateway Address of the network e.g. 11.0.1.1")
	flagSet.StringVar(&opts.ipAddr,
        "ip-address",
        "auto",
        "IP address associated with the endpoint")
	flagSet.StringVar(&opts.contName,
        "container-id",
        "",
        "Container Id to identify a runningcontainer")
	flagSet.StringVar(&opts.vlans,
        "vlans",
        "",
        "Allowed vlan ranges for auto-allocating vlans e.g. '10-100, 150-200")
	flagSet.UintVar(&opts.allocSubnetLen,
        "alloc-subnet-len",
        24,
        "Subnet length of auto allocated subnets from the subnet pool")
	flagSet.StringVar(&opts.homingHost,
        "host",
        "",
        "Host name/label on which an ep needs to be created. Default is the local host ")
	flagSet.StringVar(&opts.vxlans,
        "vxlans",
        "",
        "Allowed vlan ranges for auto-allocating vxlans e.g. '10000-20000, 30000-35000")
    flagSet.BoolVar(&opts.help, "help", false, "prints this message")
}

func usage() {
    fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", os.Args[0])
    flagSet.PrintDefaults()
}

func logFatalSubnetAndMaskFormatError() {
    log.Fatalf("gateway IP and mask must be specified e.g. 11.0.1.1/24 or " +
        "if gateway is not required to be specified then 0/24")
}

func validateOpts() error {
    var err error

	if flagSet.NArg() != 1 || opts.help {
        usage()
        os.Exit(0)
	}

	if opts.oper.Get() == "" {
		log.Fatalf("An operation must be specified")
	}

	if opts.construct.Get() == "" {
		log.Fatalf("A construct must be specified")
	}

    // global create params validation
    if opts.oper.Get() == CLI_OPER_CREATE &&
       opts.construct.Get() == CLI_CONSTRUCT_GLOBAL {
        if opts.pktTag == "vxlan" {
            log.Fatalf("default vxlan tunneling support is coming soon...")
        } else if opts.pktTagType != "vlan" {
            log.Fatalf("error '%s' packet tag type not supported", opts.pktTagType)
        } 

        _, err = netutils.ParseTagRanges(opts.vlans, "vlan")
        if err != nil {
            log.Fatalf("error '%s' parsing vlan range '%s' \n", err, opts.vlans)
        }
        _, err = netutils.ParseTagRanges(opts.vxlans, "vxlan")
        if err != nil {
            log.Fatalf("error '%s' parsing vxlan range '%s' \n", err, opts.vxlans)
        }
    }

    if opts.pktTag == "auto" {
        if opts.oper.Get() == CLI_OPER_CREATE && 
           opts.construct.Get() == CLI_CONSTRUCT_NW {
            log.Printf("Doing auto allocation of network subnet from global pool")
        }
    } else if opts.pktTag != "" {
        _, err = strconv.Atoi(opts.pktTag)
        if err != nil {
            log.Fatalf("Error convertinng tag %s to integer \n", opts.pktTag)
        }
    }

    if opts.pktTagType == "vxlan" {
        log.Fatalf("vxlan tunneling support is coming soon...")
    } else if opts.pktTagType != "vlan" {
        log.Fatalf("error '%s' packet tag type not supported", opts.pktTagType)
    } 

    // network create params validation
	if opts.oper.Get() == CLI_OPER_CREATE &&
       opts.construct.Get() == CLI_CONSTRUCT_NW {
	}

    if opts.homingHost == "" {
        opts.homingHost, err = os.Hostname()
        if err != nil {
            log.Fatalf("error obtaining the hostname, error %s \n", err)
        }
    }

    // default gw and mask parsing
    if opts.subnetCidr == "" {
        opts.subnetLen = 0
        opts.subnetIp = "auto"
    } else {
        _, _, err = net.ParseCIDR(opts.subnetCidr)
        if err != nil {
            log.Fatalf("error '%s' parsing cidr ip %s \n", err, opts.subnetCidr)
        }

        strs := strings.Split(opts.subnetCidr, "/")
        if len(strs) != 2 {
            logFatalSubnetAndMaskFormatError()
        }

        if strs[0] != "0" {
            opts.subnetIp = strs[0]
        }
        subnetLen, _ := strconv.Atoi(strs[1])
        if subnetLen > 32 {
            log.Printf("invalid mask in gateway/mask specification ")
            logFatalSubnetAndMaskFormatError()
        }
        opts.subnetLen = uint(subnetLen)
    }

	// endpoint parameters validation
	if opts.oper.Get() == CLI_OPER_CREATE &&
       opts.construct.Get() == CLI_CONSTRUCT_EP &&
       (opts.netId == "" || opts.ipAddr == "") {
        if opts.ipAddr == "auto" {
            log.Printf("doing auto ip address assignemt for the ep... \n")
        } else {
            log.Fatalf("Endpoint creation requires a valid net-id, vlan tag, " +
                "and ip address")
        }
	}

    // attach detach parameters validation
	if (opts.oper.Get() == CLI_OPER_ATTACH || opts.oper.Get() == CLI_OPER_DETACH) &&
        opts.construct.Get() == CLI_CONSTRUCT_EP && opts.contName == "" {
		log.Fatalf("A valid container-id is needed to attach/detach a container to an ep")
	}

    return err
}

func main() {
    var err error
	var state core.State = nil

	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse command. Error: %s", err)
	}
    opts.idStr = flagSet.Arg(0)
    err = validateOpts()
    if err != nil {
        os.Exit(1)
    }
    // log.Printf("parsed all valuees = %v \n", opts)

    // initialize drivers
	etcdDriver := &drivers.EtcdStateDriver{}
	driverConfig := &drivers.EtcdStateDriverConfig{}
	driverConfig.Etcd.Machines = []string{opts.etcdUrl}
	config := &core.Config{V: driverConfig}
	err = etcdDriver.Init(config)
	if err != nil {
		log.Fatalf("Failed to init etcd driver. Error: %s", err)
	}

	switch opts.construct.Get() {
	case CLI_CONSTRUCT_EP:
		if opts.oper.Get() == CLI_OPER_GET {
			epOper := &drivers.OvsOperEndpointState{StateDriver: etcdDriver}
			state = epOper
        } else if opts.oper.Get() == CLI_OPER_ATTACH || opts.oper.Get() == CLI_OPER_DETACH {
            epCfg := &drivers.OvsCfgEndpointState{StateDriver: etcdDriver}
		    err = epCfg.Read(opts.idStr)
            if err != nil {
                log.Fatalf("Failed to read ep %s. Error: %s", opts.construct.Get(), err)
            }
            log.Printf("read ep state as %v for container %s \n", epCfg, opts.contName)
            if (opts.oper.Get() == CLI_OPER_ATTACH) {
                epCfg.ContName = opts.contName
            } else {
                if epCfg.ContName != opts.contName {
                    log.Fatalf("Can not detach container '%s' from endpoint '%s' - " +
                               "container not attached \n", opts.contName, opts.idStr)
                }
                epCfg.ContName = ""
            }
            state = epCfg
        } else {
            epCfg := &drivers.OvsCfgEndpointState{StateDriver: etcdDriver}
            epCfg.Id = opts.idStr
            epCfg.NetId = opts.netId
            epCfg.IpAddress = opts.ipAddr
            epCfg.ContName = opts.contName
            epCfg.HomingHost = opts.homingHost
            state = epCfg
		}
	case CLI_CONSTRUCT_NW:
		if opts.oper.Get() == CLI_OPER_GET {
			nwOper := &drivers.OvsOperNetworkState{StateDriver: etcdDriver}
			state = nwOper
		} else {
			nwCfg := &drivers.OvsCfgNetworkState{StateDriver: etcdDriver}
            nwCfg.PktTag = opts.pktTag
            nwCfg.PktTagType = opts.pktTagType
            nwCfg.SubnetIp = opts.subnetIp
            nwCfg.SubnetLen = opts.subnetLen
            nwCfg.DefaultGw = opts.defaultGw
			nwCfg.Id = opts.idStr
			state = nwCfg
		}
    case CLI_CONSTRUCT_GLOBAL:
        var gcfg gstate.Cfg
        if opts.oper.Get() == CLI_OPER_GET {
            err = gcfg.Read(etcdDriver)
            log.Printf("State: %v \n", gcfg)
        } else {
            gcfg.Version = gstate.VersionBeta1
            gcfg.Deploy.DefaultNetType = opts.pktTagType
            gcfg.Auto.SubnetPool = opts.subnetIp
            gcfg.Auto.SubnetLen = opts.subnetLen
            gcfg.Auto.Vlans = opts.vlans
            gcfg.Auto.Vxlans = opts.vxlans
            gcfg.Auto.AllocSubnetLen = opts.allocSubnetLen
            err = gcfg.Update(etcdDriver)
        }
        if err != nil {
            log.Fatalf("error '%s' \n", err)
        }
        os.Exit(0)
	}

	switch opts.oper.Get() {
	case CLI_OPER_GET:
		err = state.Read(opts.idStr)
		if err != nil {
			log.Fatalf("Failed to read %s. Error: %s", opts.construct.Get(), err)
		} else {
			log.Printf("%s State: %v", opts.construct.Get(), state)
		}
    case CLI_OPER_ATTACH, CLI_OPER_DETACH, CLI_OPER_CREATE:
		err = state.Write()
		if err != nil {
			log.Fatalf("Failed to create %s. Error: %s", opts.construct.Get(), err)
		}
	case CLI_OPER_DELETE:
		err = state.Clear()
		if err != nil {
			log.Fatalf("Failed to delete %s. Error: %s", opts.construct.Get(), err)
		}
	}

	os.Exit(0)
}
