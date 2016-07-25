package systemtests

import (
	"flag"
	"fmt"
	"os"
	. "testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	"github.com/contiv/remotessh"
	. "gopkg.in/check.v1"
	"strings"
	. "testing"
)

type systemtestSuite struct {
	vagrant      remotessh.Vagrant
	baremetal    remotessh.Baremetal
	cli          *client.ContivClient
	short        bool
	containers   int
	binpath      string
	iterations   int
	vlanIf       string
	nodes        []*node
	fwdMode      string
	clusterStore string
	//enableDNS    bool
	//keyFile string
	//scheduler    string
	// user       string
	// password   string
	// nodes      []string
	basicInfo BasicInfo
	infoHost  InfoHost
	infoGlob  InfoGlob
}
type BasicInfo struct {
	Scheduler    string `json:"scheduler"`      //swarm, k8s or plain docker
	SwarmEnv     string `json:"swarm_variable"` //env variables to be set with swarm 
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
		if mastbasic.Scheduler == "swarm" {
			logrus.Infof("Setting up swarm cluster")
			sts.NetDemoInstallation()
		}
	}
	flag.Parse()
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
	}

	s.cli, _ = client.NewContivClient("http://localhost:9999")
}

func (s *systemtestSuite) SetUpTest(c *C) {
	logrus.Infof("============================= %s starting ==========================", c.TestName())
	outChan := make(chan string, 100)
	out, _ := s.nodes[0].runCommand("echo $DOCKER_HOST")
	outChan <- out
	logrus.Infof("DOCKER HOST for first node: %s", strings.TrimSpace(<-outChan))

	out, _ = s.nodes[1].runCommand("echo $DOCKER_HOST")
	outChan <- out
	logrus.Infof("DOCKER HOST for second node: %s", strings.TrimSpace(<-outChan))

	out, _ = s.nodes[0].runCommand("docker info")
	outChan <- out
	logrus.Infof("DOCKER INFO for first node: %s", strings.TrimSpace(<-outChan))

	out, _ = s.nodes[1].runCommand("docker info")
	outChan <- out
	logrus.Infof("DOCKER INFO for second node: %s", strings.TrimSpace(<-outChan))
	switch s.basicInfo.Platform {
	case "baremetal":
		s.SetUpTestBaremetal(c)

	case "vagrant":
		s.SetUpTestVagrant(c)

	}

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
	logrus.Infof("-----in test00SSH-------")
	switch s.basicInfo.Platform {
	case "baremetal":
		c.Assert(s.baremetal.IterateNodes(func(node vagrantssh.TestbedNode) error {
			return node.RunCommand("true")
		}), IsNil)

	case "vagrant":
		c.Assert(s.vagrant.IterateNodes(func(node vagrantssh.TestbedNode) error {
			return node.RunCommand("true")
		}), IsNil)

	}
}
