package systemtests

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
	"os"

	"strings"
	. "testing"
	"time"
)

type systemtestSuite struct {
	vagrant   vagrantssh.Vagrant
	baremetal vagrantssh.Baremetal
	cli       *client.ContivClient
	nodes     []*node
	fwdMode   string
	basicInfo BasicInfo
	infoHost  InfoHost
	infoGlob  InfoGlob
}
type BasicInfo struct {
	Scheduler    string `json:"scheduler"`      //swarm, k8s or plain docker
	SwarmEnv     string `json:"swarm_variable"` //env variables to be set with swarm environment
	Platform     string `json:"platform"`       //vagrant or baremetal
	Product      string `json:"product"`        //for netplugin / volplugin
	AciMode      string `json:"aci_mode"`       //on/off
	Short        bool   `json:"short"`
	Containers   int    `json:"containers"`
	Iterations   int    `json:"iterations"`
	EnableDNS    bool   `json:"enableDNS"`
	ClusterStore string `json:"contiv_cluster_store"`
	ContivL3     string `json:"contiv_l3"`
	KeyFile      string `json:"key_file"`
	BinPath      string `json:"binpath"` // /home/admin/bin or /opt/gopath/bin
	Master       bool   `json:"master"`
}

type InfoHost struct {
	HostIPs           string `json:"hostips"`
	HostUsernames     string `json:"hostusernames"`
	HostDataInterface string `json:"dataInterface"`
	HostMgmtInterface string `json:"mgmtInterface"`
	Master            bool   `json:"master"`
}

type InfoGlob struct {
	Vlan    string `json:"vlan"`
	Vxlan   string `json:"vxlan"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	Network string `json:"network"`
	Tenant  string `json:"tenant"`
	Encap   string `json:"encap"`
	Master  bool   `json:"master"`
}

var sts = &systemtestSuite{}

var _ = Suite(sts)

func TestMain(m *M) {

	mastbasic, _, _ := getMaster("cfg.json")

	if mastbasic.ContivL3 == "" {
		flag.StringVar(&sts.fwdMode, "fwd-mode", "bridge", "forwarding mode to start the test ")
	} else {
		flag.StringVar(&sts.fwdMode, "fwd-mode", "routing", "forwarding mode to start the test ")
	}
	if mastbasic.Platform == "baremetal" {
		logrus.Infof("Starting net_demo_installer")
		sts.BaremetalSetup()
	}
	flag.Parse()
	//logrus.Infof("Running system test with params: %+v", sts)
	os.Exit(m.Run())
}

func TestSystem(t *T) {
	if os.Getenv("HOST_TEST") != "" {
		os.Exit(0)
	}

	TestingT(t)
}

func (s *systemtestSuite) SetUpSuite(c *C) {
	logrus.Infof("Bootstrapping system tests")
	s.basicInfo, s.infoHost, s.infoGlob = getMaster("cfg.json")

	switch s.basicInfo.Platform {
	case "baremetal":
		s.SetUpSuiteBaremetal(c)

	case "vagrant":
		s.SetUpSuiteVagrant(c)
	} // end of switch case

	s.cli, _ = client.NewContivClient("http://localhost:9999")
}

func (s *systemtestSuite) SetUpTest(c *C) {
	logrus.Infof("============================= %s starting ==========================", c.TestName())

	switch s.basicInfo.Platform {

	case "baremetal":
		logrus.Infof("-----Inside  switch case ------")
		for _, node := range s.nodes {
			node.exec.cleanupContainers()
			node.exec.cleanupDockerNetwork()

			node.stopNetplugin()
			node.cleanupSlave()
			node.deleteFile("/etc/systemd/system/netplugin.service")
			node.stopNetmaster()
			node.deleteFile("/etc/systemd/system/netmaster.service")
			node.deleteFile("/usr/bin/netplugin")
			node.deleteFile("/usr/bin/netmaster")
			node.deleteFile("/usr/bin/netctl")
		}

		for _, node := range s.nodes {
			node.cleanupMaster()
		}

		for _, node := range s.nodes {
			if s.fwdMode == "bridge" {
				c.Assert(node.startNetplugin(""), IsNil)
				time.Sleep(15 * time.Second)
				c.Assert(node.exec.runCommandUntilNoNetpluginError(), IsNil)
			} else if s.fwdMode == "routing" {
				c.Assert(node.startNetplugin("-fwd-mode=routing -vlan-if=eth2"), IsNil)
				c.Assert(node.exec.runCommandUntilNoNetpluginError(), IsNil)
			}
		}

		time.Sleep(15 * time.Second)

		for _, node := range s.nodes {
			c.Assert(node.startNetmaster(), IsNil)
			time.Sleep(1 * time.Second)
			c.Assert(node.exec.runCommandUntilNoNetmasterError(), IsNil)
		}

		time.Sleep(5 * time.Second)
		for i := 0; i < 11; i++ {
			_, err := s.cli.TenantGet("default")
			if err == nil {
				break
			}
			// Fail if we reached last iteration
			c.Assert((i < 10), Equals, true)
			time.Sleep(500 * time.Millisecond)
		}
	case "vagrant":
		for _, node := range s.nodes {
			node.exec.cleanupContainers()
			node.exec.cleanupDockerNetwork()
			node.stopNetplugin()
			node.cleanupSlave()
		}

		for _, node := range s.nodes {
			node.stopNetmaster()

		}
		for _, node := range s.nodes {
			node.cleanupMaster()
		}

		for _, node := range s.nodes {
			if s.fwdMode == "bridge" {
				c.Assert(node.startNetplugin(""), IsNil)
				logrus.Infof("-----CHECKPOINT 3-------")
				c.Assert(node.exec.runCommandUntilNoNetpluginError(), IsNil)
				logrus.Infof("-----CHECKPOINT 3-------")
			} else if s.fwdMode == "routing" {
				c.Assert(node.startNetplugin("-fwd-mode=routing -vlan-if=eth2"), IsNil)
				c.Assert(node.exec.runCommandUntilNoNetpluginError(), IsNil)
			}
		}

		time.Sleep(15 * time.Second)

		// temporarily enable DNS for service discovery tests
		prevDNSEnabled := s.basicInfo.EnableDNS
		if strings.Contains(c.TestName(), "SvcDiscovery") {
			s.basicInfo.EnableDNS = true
		}

		defer func() { s.basicInfo.EnableDNS = prevDNSEnabled }()

		for _, node := range s.nodes {
			c.Assert(node.startNetmaster(), IsNil)
			time.Sleep(1 * time.Second)
			c.Assert(node.exec.runCommandUntilNoNetmasterError(), IsNil)
		}

		time.Sleep(5 * time.Second)
		if s.basicInfo.Scheduler != "k8" {
			for i := 0; i < 11; i++ {

				_, err := s.cli.TenantGet("default")
				if err == nil {
					break
				}
				// Fail if we reached last iteration
				c.Assert((i < 10), Equals, true)
				time.Sleep(500 * time.Millisecond)
			}
		}
	} // end of switch case

}

func (s *systemtestSuite) TearDownTest(c *C) {
	for _, node := range s.nodes {
		c.Check(node.checkForNetpluginErrors(), IsNil)
		c.Assert(node.exec.rotateNetpluginLog(), IsNil)
		c.Assert(node.exec.rotateNetmasterLog(), IsNil)
	}
	logrus.Infof("============================= %s completed ==========================", c.TestName())
}

func (s *systemtestSuite) TearDownSuite(c *C) {

	for _, node := range s.nodes {
		node.exec.cleanupContainers()
	}

	// Print all errors and fatal messages
	for _, node := range s.nodes {
		logrus.Infof("Checking for errors on %v", node.Name())
		out, _ := node.runCommand(`for i in /tmp/_net*; do grep "error\|fatal\|panic" $i; done`)
		if out != "" {
			logrus.Errorf("Errors in logfiles on %s: \n", node.Name())
			fmt.Printf("%s\n==========================\n\n", out)
		}
	}
}

func (s *systemtestSuite) Test00SSH(c *C) {
	c.Assert(s.baremetal.IterateNodes(func(node vagrantssh.TestbedNode) error {
		logrus.Infof("-----in test00SSH-------")
		return node.RunCommand("true")
	}), IsNil)
}
