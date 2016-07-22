package systemtests

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
	"os"
	"os/exec"
	"strconv"
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
	IP                string `json:"ip"`
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

	logrus.Infof("keyfle value is %s", mastbasic.KeyFile)
	logrus.Infof("binpath value is %s", mastbasic.BinPath)
	//logrus.Infof("vlanif is %s", mastbasic.VlanIf)

	if mastbasic.ContivL3 == "" {
		flag.StringVar(&sts.fwdMode, "fwd-mode", "bridge", "forwarding mode to start the test ")
	} else {
		flag.StringVar(&sts.fwdMode, "fwd-mode", "routing", "forwarding mode to start the test ")
	}
	if mastbasic.Platform == "Baremetal" {
		logrus.Infof("cmae here")
		//sts.BaremetalSetup()
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
	case "Baremetal":
		logrus.Infof("ACI_SYS_TEST_MODE is on")
		logrus.Infof("Private keyFile = %s", s.basicInfo.KeyFile)
		logrus.Infof("Binary binpath = %s", s.basicInfo.BinPath)
		logrus.Infof("Interface vlanIf = %s", s.infoHost.HostDataInterface)

		s.baremetal = vagrantssh.Baremetal{}
		bm := &s.baremetal

		// To fill the hostInfo data structure for Baremetal VMs
		name := "aci-swarm-node"
		hostIPs := strings.Split(s.infoHost.HostIPs, ",")
		hostNames := strings.Split(s.infoHost.HostUsernames, ",")
		hosts := make([]vagrantssh.HostInfo, 2)

		for i := range hostIPs {
			hosts[i].Name = name + strconv.Itoa(i+1)
			logrus.Infof("Name=%s", hosts[i].Name)

			hosts[i].SSHAddr = hostIPs[i]
			logrus.Infof("SHAddr=%s", hosts[i].SSHAddr)

			hosts[i].SSHPort = "22"

			hosts[i].User = hostNames[i]
			logrus.Infof("User=%s", hosts[i].User)

			hosts[i].PrivKeyFile = s.basicInfo.KeyFile
			logrus.Infof("PrivKeyFile=%s", hosts[i].PrivKeyFile)
		}
		logrus.Infof("hosts are %s", hosts)
		c.Assert(bm.Setup(hosts), IsNil)

		s.nodes = []*node{}

		for _, nodeObj := range s.baremetal.GetNodes() {
			logrus.Infof("node name is %s", nodeObj.GetName())
			nodeName := nodeObj.GetName()
			if strings.Contains(nodeName, "aci") ||
				strings.Contains(nodeName, "swarm") {
				node := &node{}
				node.tbnode = nodeObj
				node.suite = s

				switch s.basicInfo.Scheduler {
				case "k8":
					node.exec = s.NewK8sExec(node)
				case "swarm":
					logrus.Infof("#############in swarm")
					node.exec = s.NewSwarmExec(node)
				default:
					logrus.Infof("in docker MOOOOOD")
					node.exec = s.NewDockerExec(node)
				}
				s.nodes = append(s.nodes, node)
			}
			//s.nodes = append(s.nodes, &node{tbnode: nodeObj, suite: s})
		}

		logrus.Info("Pulling alpine on all nodes")

		s.baremetal.IterateNodes(func(node vagrantssh.TestbedNode) error {
			//fmt.Printf("\n\t Node Name is %s", node.nodeName)
			node.RunCommand("sudo rm /tmp/*net*")
			node.RunCommand("touch /home/admin/GAURAV.txt")
			return node.RunCommand("docker pull alpine")
		})
		s.BaremetalTestInstall(c)
		//Copying binaries
		s.copyBinary("netmaster")
		s.copyBinary("netplugin")
		s.copyBinary("netctl")
		s.copyBinary("contivk8s")

	case "Vagrant":
		s.vagrant = vagrantssh.Vagrant{}
		nodesStr := os.Getenv("CONTIV_NODES")
		var contivNodes int

		if nodesStr == "" {
			contivNodes = 2
		} else {
			var err error
			contivNodes, err = strconv.Atoi(nodesStr)
			if err != nil {
				c.Fatal(err)
			}
		}

		s.nodes = []*node{}

		outChan := make(chan string, 100)
		//logrus.Infof("env value is " + s.basicInfo.SwarmEnv)

		mystr := "DOCKER_HOST=10.193.246.70:2375 docker info"
		logrus.Infof("mystr _____________________ value is " + mystr)
		out, _ := s.nodes[0].runCommand(mystr)
		outChan <- out
		logrus.Infof("docker ps for first node ====== %s", strings.TrimSpace(<-outChan))

		if s.fwdMode == "routing" {
			contivL3Nodes := 2
			switch s.basicInfo.Scheduler {
			case "k8":
				contivNodes = 4
				c.Assert(s.vagrant.Setup(false, []string{"CONTIV_L3=1 VAGRANT_CWD=/home/ladmin/src/github.com/contiv/netplugin/vagrant/k8s/"}, contivNodes), IsNil)
			case "swarm":
				c.Assert(s.vagrant.Setup(false, append([]string{"CONTIV_NODES=3 CONTIV_L3=1"}, s.basicInfo.SwarmEnv), contivNodes+contivL3Nodes), IsNil)
			default:
				c.Assert(s.vagrant.Setup(false, []string{"CONTIV_NODES=3 CONTIV_L3=1"}, contivNodes+contivL3Nodes), IsNil)

			}

		} else {
			switch s.basicInfo.Scheduler {
			case "k8":
				contivNodes = contivNodes + 1 //k8master
				c.Assert(s.vagrant.Setup(false, []string{"VAGRANT_CWD=/home/ladmin/src/github.com/contiv/netplugin/vagrant/k8s/"}, contivNodes), IsNil)
			case "swarm":
				c.Assert(s.vagrant.Setup(false, append([]string{}, s.basicInfo.SwarmEnv), contivNodes), IsNil)
			default:
				c.Assert(s.vagrant.Setup(false, []string{}, contivNodes), IsNil)

			}

		}
		logrus.Infof("Checkpoint 1-----")
		for _, nodeObj := range s.vagrant.GetNodes() {
			nodeName := nodeObj.GetName()
			if strings.Contains(nodeName, "netplugin-node") ||
				strings.Contains(nodeName, "k8") {
				node := &node{}
				node.tbnode = nodeObj
				node.suite = s
				logrus.Infof("scheduler is %s ", s.basicInfo.Scheduler)
				switch s.basicInfo.Scheduler {
				case "k8":
					node.exec = s.NewK8sExec(node)
				case "swarm":
					logrus.Infof("in swarm mooooood")
					node.exec = s.NewSwarmExec(node)
				default:
					logrus.Infof("in docker mooooood")
					node.exec = s.NewDockerExec(node)
				}
				s.nodes = append(s.nodes, node)
			}
		}
		logrus.Infof("Checkpoint 2-----")
		logrus.Info("Pulling alpine on all nodes")
		s.vagrant.IterateNodes(func(node vagrantssh.TestbedNode) error {
			node.RunCommand("sudo rm /tmp/net*")
			return node.RunCommand("docker pull alpine")
		})
	} // end of switch case

	s.cli, _ = client.NewContivClient("http://localhost:9999")
}

func (s *systemtestSuite) SetUpTest(c *C) {
	logrus.Infof("============================= %s starting ==========================", c.TestName())

	switch s.basicInfo.Platform {

	case "Baremetal":
		logrus.Infof("-----Inside  switch case ------")
		for _, node := range s.nodes {
			//node.exec.cleanupContainers()
			//node.exec.cleanupDockerNetwork()

			node.stopNetplugin()
			//node.cleanupSlave()
			node.deleteFile("/etc/systemd/system/netplugin.service")
			node.stopNetmaster()
			node.deleteFile("/etc/systemd/system/netmaster.service")
			//node.deleteFile("/usr/bin/netplugin")
			//node.deleteFile("/usr/bin/netmaster")
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
	case "Vagrant":
		for _, node := range s.nodes {
			node.exec.cleanupContainers()
			//node.cleanupDockerNetwork()
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

func (s *systemtestSuite) BaremetalSetup() {
	cmd := exec.Command("chmod +x", "net_demo_installer")
	cmd.Run()
	cmd = exec.Command("./net_demo_installer", "-ar")
	// setup log file
	file, err := os.Create("server.log")
	if err != nil {
		logrus.Infof("no err here")
	}
	cmd.Stdout = file

	cmd.Run()
	logrus.Infof("Done running net demo ------------------")
}

func (s *systemtestSuite) BaremetalTestInstall(c *C) {
	outChan := make(chan string, 100)
	mystr := "docker info | grep Nodes"
	out := ""
	i := 1
	out1 := ""

	err := ""

	for _, node := range s.nodes {

		if i == 1 {
			out1, _ = node.runCommand(mystr)
			outChan <- out1
			logrus.Infof("for first node docker info | grep nodes ====== %s", strings.TrimSpace(<-outChan))

			if out1 == "" {
				logrus.Infof("Nothing found on the first node ")
				err = "net_demo didnt run"
				break
			}
		} else {
			out, _ = node.runCommand(mystr)

			outChan <- out
			logrus.Infof("docker info | grep nodes ====== %s", strings.TrimSpace(<-outChan))
			if out != out1 {
				logrus.Infof("Nodes not in sync")
				err = "net_demo didnt run"
				break
			}
		}
	}
	cmd := exec.Command("sudo rm -rf", "ansible genInventoryFile.py server.log")
	cmd.Run()
	c.Assert(err, Equals, "")
}
