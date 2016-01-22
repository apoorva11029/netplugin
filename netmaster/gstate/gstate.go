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

package gstate

import (
	"encoding/json"

	"github.com/jainvipin/bitset"

	"github.com/contiv/netplugin/core"
	"github.com/contiv/netplugin/netmaster/resources"
	"github.com/contiv/netplugin/utils/netutils"

	log "github.com/Sirupsen/logrus"
)

const (
	baseGlobal          = "/contiv.io/"
	cfgGlobalPrefix     = baseGlobal + "config/global/"
	cfgGlobalPath       = cfgGlobalPrefix + "global"
	operGlobalPrefix    = baseGlobal + "oper/global/"
	operGlobalPath      = operGlobalPrefix + "global"
	vxlanLocalVlanRange = "1-4094"
)

// Version constants. Used in managing state variance.
const (
	VersionBeta1 = "0.01"
)

// AutoParams specifies various parameters for the auto allocation and resource
// management for networks and endpoints.  This allows for hands-free
// allocation of resources without having to specify these each time these
// constructs gets created.
type AutoParams struct {
	VLANs  string `json:"VLANs"`
	VXLANs string `json:"VXLANs"`
}

// Cfg is the configuration of a tenant.
type Cfg struct {
	core.CommonState
	Version string     `json:"version"`
	Auto    AutoParams `json:"auto"`
}

// Oper encapsulates operations on a tenant.
type Oper struct {
	core.CommonState
	DefaultNetwork  string `json:"defaultNetwork"`
	FreeVXLANsStart uint   `json:"freeVXLANsStart"`
}

// Dump is a debugging utility.
func (gc *Cfg) Dump() error {
	log.Debugf("Global State %v \n", gc)
	return nil
}

func (gc *Cfg) checkErrors() error {
	var err error

	_, err = netutils.ParseTagRanges(gc.Auto.VLANs, "vlan")
	if err != nil {
		return err
	}

	_, err = netutils.ParseTagRanges(gc.Auto.VXLANs, "vxlan")
	if err != nil {
		return err
	}

	return err
}

// Parse parses a JSON config into a *gstate.Cfg.
func Parse(configBytes []byte) (*Cfg, error) {
	var gc Cfg

	err := json.Unmarshal(configBytes, &gc)
	if err != nil {
		return nil, err
	}

	err = gc.checkErrors()
	if err != nil {
		return nil, err
	}

	return &gc, err
}

// Write the state
func (gc *Cfg) Write() error {
	key := cfgGlobalPath
	return gc.StateDriver.WriteState(key, gc, json.Marshal)
}

// Read the state
func (gc *Cfg) Read(dummy string) error {
	key := cfgGlobalPath
	return gc.StateDriver.ReadState(key, gc, json.Unmarshal)
}

// ReadAll global config state
func (gc *Cfg) ReadAll() ([]core.State, error) {
	return gc.StateDriver.ReadAllState(cfgGlobalPrefix, gc, json.Unmarshal)
}

// Clear the state
func (gc *Cfg) Clear() error {
	key := cfgGlobalPath
	return gc.StateDriver.ClearState(key)
}

// Write the state
func (g *Oper) Write() error {
	key := operGlobalPath
	return g.StateDriver.WriteState(key, g, json.Marshal)
}

// Read the state
func (g *Oper) Read(dummy string) error {
	key := operGlobalPath
	return g.StateDriver.ReadState(key, g, json.Unmarshal)
}

// ReadAll the global oper state
func (g *Oper) ReadAll() ([]core.State, error) {
	return g.StateDriver.ReadAllState(operGlobalPrefix, g, json.Unmarshal)
}

// Clear the state.
func (g *Oper) Clear() error {
	key := operGlobalPath
	return g.StateDriver.ClearState(key)
}

func (gc *Cfg) initVXLANBitset(vxlans string) (*resources.AutoVXLANCfgResource, uint, error) {

	vxlanRsrcCfg := &resources.AutoVXLANCfgResource{}
	vxlanRsrcCfg.VXLANs = netutils.CreateBitset(14)

	vxlanRange := netutils.TagRange{}
	vxlanRanges, err := netutils.ParseTagRanges(vxlans, "vxlan")
	if err != nil {
		return nil, 0, err
	}
	// XXX: REVISIT, we seem to accept one contiguous vxlan range
	vxlanRange = vxlanRanges[0]

	freeVXLANsStart := uint(vxlanRange.Min) - 1
	for vxlan := vxlanRange.Min; vxlan <= vxlanRange.Max; vxlan++ {
		vxlanRsrcCfg.VXLANs.Set(uint(vxlan) - freeVXLANsStart)
	}

	// Initialize local vlan bitset
	vxlanRsrcCfg.LocalVLANs, err = gc.initVLANBitset(vxlanLocalVlanRange)
	if err != nil {
		return nil, 0, err
	}

	return vxlanRsrcCfg, freeVXLANsStart, nil
}

// AllocVXLAN allocates a new vxlan; ids for both the vxlan and vlan are returned.
func (gc *Cfg) AllocVXLAN(reqVxlan uint) (vxlan uint, localVLAN uint, err error) {

	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return 0, 0, err
	}
	ra := core.ResourceManager(tempRm)

	g := &Oper{}
	g.StateDriver = gc.StateDriver
	err = g.Read("")
	if err != nil {
		return 0, 0, err
	}

	if (reqVxlan != 0) && (reqVxlan >= g.FreeVXLANsStart) {
		reqVxlan = reqVxlan - g.FreeVXLANsStart
	}

	pair, err1 := ra.AllocateResourceVal("global", resources.AutoVXLANResource, reqVxlan)
	if err1 != nil {
		return 0, 0, err1
	}

	vxlan = pair.(resources.VXLANVLANPair).VXLAN + g.FreeVXLANsStart
	localVLAN = pair.(resources.VXLANVLANPair).VLAN

	return
}

// FreeVXLAN returns a VXLAN id to the pool.
func (gc *Cfg) FreeVXLAN(vxlan uint, localVLAN uint) error {
	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return err
	}
	ra := core.ResourceManager(tempRm)

	g := &Oper{}
	g.StateDriver = gc.StateDriver
	err = g.Read("")
	if err != nil {
		return nil
	}

	return ra.DeallocateResourceVal("global", resources.AutoVXLANResource,
		resources.VXLANVLANPair{
			VXLAN: vxlan - g.FreeVXLANsStart,
			VLAN:  localVLAN})
}

func clearReservedVLANs(vlanBitset *bitset.BitSet) {
	vlanBitset.Clear(0)
	vlanBitset.Clear(4095)
}

func (gc *Cfg) initVLANBitset(vlans string) (*bitset.BitSet, error) {

	vlanBitset := netutils.CreateBitset(12)

	vlanRanges, err := netutils.ParseTagRanges(vlans, "vlan")
	if err != nil {
		return nil, err
	}

	for _, vlanRange := range vlanRanges {
		for vlan := vlanRange.Min; vlan <= vlanRange.Max; vlan++ {
			vlanBitset.Set(uint(vlan))
		}
	}
	clearReservedVLANs(vlanBitset)

	return vlanBitset, nil
}

// AllocVLAN allocates a new VLAN resource. Returns an ID.
func (gc *Cfg) AllocVLAN(reqVlan uint) (uint, error) {
	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return 0, err
	}
	ra := core.ResourceManager(tempRm)

	vlan, err := ra.AllocateResourceVal("global", resources.AutoVLANResource, reqVlan)
	if err != nil {
		log.Errorf("alloc vlan failed: %q", err)
		return 0, err
	}

	return vlan.(uint), err
}

// FreeVLAN releases a VLAN for a given ID.
func (gc *Cfg) FreeVLAN(vlan uint) error {
	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return err
	}
	ra := core.ResourceManager(tempRm)

	return ra.DeallocateResourceVal("global", resources.AutoVLANResource, vlan)
}

// Process validates, implements, and writes the state.
func (gc *Cfg) Process() error {
	var err error

	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return err
	}

	ra := core.ResourceManager(tempRm)

	if gc.Version != VersionBeta1 {
		return core.Errorf("unsupported version %s", gc.Version)
	}

	err = gc.checkErrors()
	if err != nil {
		return core.Errorf("process failed on error checks %s", err)
	}

	// Only define a vlan resource if a valid range was specified
	if gc.Auto.VLANs != "" {
		var vlanRsrcCfg *bitset.BitSet
		vlanRsrcCfg, err = gc.initVLANBitset(gc.Auto.VLANs)
		if err != nil {
			return err
		}
		err = ra.DefineResource("global", resources.AutoVLANResource, vlanRsrcCfg)
		if err != nil {
			return err
		}
	}

	// Only define a vxlan resource if a valid range was specified
	var freeVXLANsStart uint
	if gc.Auto.VXLANs != "" {
		var vxlanRsrcCfg *resources.AutoVXLANCfgResource
		vxlanRsrcCfg, freeVXLANsStart, err = gc.initVXLANBitset(gc.Auto.VXLANs)
		if err != nil {
			return err
		}
		err = ra.DefineResource("global", resources.AutoVXLANResource, vxlanRsrcCfg)
		if err != nil {
			return err
		}
	}

	g := &Oper{FreeVXLANsStart: freeVXLANsStart}
	g.StateDriver = gc.StateDriver
	err = g.Write()
	if err != nil {
		log.Errorf("error '%s' updating goper state %v \n", err, g)
		return err
	}

	log.Debugf("updating the global config to new state %v \n", gc)
	return nil
}

// DeleteResources deletes associated resources
func (gc *Cfg) DeleteResources() error {
	tempRm, err := resources.GetStateResourceManager()
	if err != nil {
		return err
	}

	ra := core.ResourceManager(tempRm)

	err = ra.UndefineResource("global", resources.AutoVLANResource)
	if err != nil {
		log.Errorf("Error deleting vlan resource. Err: %v", err)
	}

	err = ra.UndefineResource("global", resources.AutoVXLANResource)
	if err != nil {
		log.Errorf("Error deleting vxlan resource. Err: %v", err)
	}

	return err
}

// AssignDefaultNetwork assigns a default network for a tenant based on the configuration
// in case configuration is absent it uses the provided network name to be the default
// network. It records the default network in oper state (derived or configured)
func (gc *Cfg) AssignDefaultNetwork(networkName string) (string, error) {
	g := &Oper{}
	g.StateDriver = gc.StateDriver
	if err := g.Read(""); core.ErrIfKeyExists(err) != nil {
		return "", err
	}
	if g.DefaultNetwork != "" {
		return "", nil
	}

	// not checking if this network exists within a tenant
	g.DefaultNetwork = networkName

	if err := g.Write(); err != nil {
		log.Errorf("error '%s' updating goper state %v \n", err, g)
		return "", err
	}

	return g.DefaultNetwork, nil
}

// UnassignNetwork clears the oper state w.r.t. default network name
func (gc *Cfg) UnassignNetwork(networkName string) error {
	if networkName == "" {
		return nil
	}

	g := &Oper{}
	g.StateDriver = gc.StateDriver
	if err := g.Read(""); core.ErrIfKeyExists(err) != nil {
		return err
	}

	if networkName == g.DefaultNetwork {
		g.DefaultNetwork = ""
		if err := g.Write(); err != nil {
			log.Errorf("error '%s' updating goper state %v \n", err, g)
			return err
		}
	}

	return nil
}
