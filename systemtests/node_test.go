package systemtests

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/systemtests-utils"
	"github.com/contiv/vagrantssh"
)

type node struct {
	tbnode vagrantssh.TestbedNode
	suite  *systemtestSuite
	exec   systemTestScheduler
}

type containerSpec struct {
	imageName   string
	commandName string
	networkName string
	serviceName string
	tenantName  string
	name        string
	dnsServer   string
	labels      []string
	epGroup     string
}

func (n *node) rotateLog(prefix string) error {
	if prefix == "netmaster" {
		return n.exec.rotateNetmasterLog()
	} else if prefix == "netplugin" {
		return n.exec.rotateNetpluginLog()
	}
	return nil
}

func (n *node) getIPAddr(dev string) (string, error) {
	out, err := n.runCommand(fmt.Sprintf("ip addr show dev %s | grep inet | head -1", dev))
	if err != nil {
		logrus.Errorf("Failed to get IP for node %v", n.tbnode)
		logrus.Println(out)
	}

	parts := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(out), -1)
	if len(parts) < 2 {
		return "", fmt.Errorf("Invalid output from node %v: %s", n.tbnode, out)
	}

	parts = strings.Split(parts[1], "/")
	out = strings.TrimSpace(parts[0])
	return out, err
}

func (n *node) Name() string {
	return n.tbnode.GetName()
}

func (s *systemtestSuite) getNodeByName(name string) *node {
	for _, myNode := range s.nodes {
		if myNode.Name() == name {
			return myNode
		}
	}

	return nil
}

func (n *node) startNetplugin(args string) error {
	logrus.Infof("Starting netplugin on %s", n.Name())
	return n.exec.startNetplugin(args)
}

func (n *node) stopNetplugin() error {
	return n.exec.stopNetplugin()
}

func (s *systemtestSuite) copyBinary(fileName string) error {
	logrus.Infof("Copying %s binary to %s", fileName, s.basicInfo.BinPath)
	hostIPs := strings.Split(os.Getenv("HOST_IPS"), ",")
	srcFile := s.basicInfo.BinPath + "/" + fileName
	destFile := s.basicInfo.BinPath + "/" + fileName
	for i := 1; i < len(s.nodes); i++ {
		logrus.Infof("Copying %s binary to IP= %s and Directory = %s", srcFile, hostIPs[i], destFile)
		s.nodes[0].tbnode.RunCommand("scp -i " + s.basicInfo.KeyFile + " " + srcFile + " " + hostIPs[i] + ":" + destFile)
	}
	return nil
}

func (n *node) deleteFile(file string) error {
	logrus.Infof("Deleting %s file ", file)
	return n.tbnode.RunCommand("sudo rm " + file)
}

func (n *node) stopNetmaster() error {
	return n.exec.stopNetmaster()
}

func (n *node) startNetmaster() error {
	return n.exec.startNetmaster()
}

func (n *node) cleanupSlave() {
	n.exec.cleanupSlave()
}

func (n *node) cleanupMaster() {
	n.exec.cleanupMaster()
}

func (n *node) runCommand(cmd string) (string, error) {
	var (
		str string
		err error
	)

	for {
		str, err = n.tbnode.RunCommandWithOutput(cmd)
		if err == nil || !strings.Contains(err.Error(), "EOF") {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return str, err
}

func (n *node) checkForNetpluginErrors() error {
	return n.exec.checkForNetpluginErrors()
}

func (n *node) runCommandUntilNoError(cmd string) error {
	runCmd := func() (string, bool) {
		if err := n.tbnode.RunCommand(cmd); err != nil {
			return "", false
		}
		return "", true
	}
	timeoutMessage := fmt.Sprintf("timeout reached trying to run %v on %q", cmd, n.Name())
	_, err := utils.WaitForDone(runCmd, 10*time.Millisecond, 10*time.Second, timeoutMessage)
	return err
}

func (n *node) checkPingWithCount(ipaddr string, count int) error {
	logrus.Infof("Checking ping from %s to %s", n.Name(), ipaddr)
	cmd := fmt.Sprintf("ping -c %d %s", count, ipaddr)
	out, err := n.tbnode.RunCommandWithOutput(cmd)

	if err != nil || strings.Contains(out, "0 received, 100% packet loss") {
		logrus.Errorf("Ping from %s to %s FAILED: %q - %v", n.Name(), ipaddr, out, err)
		return fmt.Errorf("Ping failed from %s to %s: %q - %v", n.Name(), ipaddr, out, err)
	}

	logrus.Infof("Ping from %s to %s SUCCEEDED", n.Name(), ipaddr)
	return nil
}

func (n *node) checkPing(ipaddr string) error {
	return n.checkPingWithCount(ipaddr, 1)
}

func (n *node) reloadNode() error {
	logrus.Infof("Reloading node %s", n.Name())

	out, err := exec.Command("vagrant", "reload", n.Name()).CombinedOutput()
	if err != nil {
		logrus.Errorf("Error reloading node %s. Err: %v\n Output: %s", n.Name(), err, string(out))
		return err
	}

	logrus.Infof("Reloaded node %s. Output:\n%s", n.Name(), string(out))
	return nil
}

func (n *node) restartClusterStore() error {
	if strings.Contains(n.suite.basicInfo.ClusterStore, "etcd://") {
		logrus.Infof("Restarting etcd on %s", n.Name())

		n.runCommand("sudo systemctl stop etcd")
		time.Sleep(5 * time.Second)
		n.runCommand("sudo systemctl start etcd")

		logrus.Infof("Restarted etcd on %s", n.Name())
	} else if strings.Contains(n.suite.basicInfo.ClusterStore, "consul://") {
		logrus.Infof("Restarting consul on %s", n.Name())

		n.runCommand("sudo systemctl stop consul")
		time.Sleep(5 * time.Second)
		n.runCommand("sudo systemctl start consul")

		logrus.Infof("Restarted consul on %s", n.Name())
	}

	return nil
}
