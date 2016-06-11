package systemtests

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type intf struct {
	ip   string
	ipv6 string
}

type container struct {
	node        *node
	containerID string
	name        string
	eth0        intf
}
type docker struct {
	node *node
}

func (s *systemtestSuite) NewDockerExec(n *node) *docker {
	d := new(docker)
	d.node = n
	return d
}

func (d *docker) newContainer(node *node, containerID, name string) (*container, error) {
	cont := &container{node: node, containerID: containerID, name: name}

	out, err := node.exec.getIPAddr(cont, "eth0")
	if err != nil {
		return nil, err
	}
	cont.eth0.ip = out

	out, err = cont.getIPv6Addr("eth0")
	if err == nil {
		cont.eth0.ipv6 = out
	}

	return cont, nil
}

func (d *docker) runContainer(spec containerSpec) (*container, error) {
	var namestr, netstr, dnsStr, labelstr string

	if spec.networkName != "" {
		netstr = spec.networkName
		if spec.tenantName != "" {
			netstr = netstr + "/" + spec.tenantName
		}

		if spec.serviceName != "" {
			netstr = spec.serviceName
		}

		netstr = "--net=" + netstr
	}

	if spec.imageName == "" {
		spec.imageName = "alpine"
	}

	if spec.commandName == "" {
		spec.commandName = "sleep 60m"
	}

	if spec.name != "" {
		namestr = "--name=" + spec.name
	}

	if spec.dnsServer != "" {
		dnsStr = "--dns=" + spec.dnsServer
	}

	if len(spec.labels) > 0 {
		l := "--label="
		for _, label := range spec.labels {
			labelstr += l + label + " "
		}
	}

	logrus.Infof("Starting a container running %q on %s", spec.commandName, d.node.Name())

	cmd := fmt.Sprintf("docker run -itd %s %s %s %s %s %s", namestr, netstr, dnsStr, labelstr, spec.imageName, spec.commandName)

	out, err := d.node.tbnode.RunCommandWithOutput(cmd)
	if err != nil {
		logrus.Infof("cmd %q failed: output below", cmd)
		logrus.Println(out)
		out2, err := d.node.tbnode.RunCommandWithOutput(fmt.Sprintf("docker logs %s", strings.TrimSpace(out)))
		if err == nil {
			logrus.Println(out2)
		} else {
			logrus.Errorf("Container id %q is invalid", strings.TrimSpace(out))
		}

		return nil, err
	}

	cont, err := d.newContainer(d.node, strings.TrimSpace(out), spec.name)
	if err != nil {
		logrus.Info(err)
		return nil, err
	}

	return cont, nil
}

func (d *docker) checkPingFailure(c *container, ipaddr string) error {
	logrus.Infof("Expecting ping failure from %v to %s", c, ipaddr)
	if err := d.checkPing(c, ipaddr); err == nil {
		return fmt.Errorf("Ping succeeded when expected to fail from %v to %s", c, ipaddr)
	}

	return nil
}

func (d *docker) checkPing(c *container, ipaddr string) error {
	logrus.Infof("Checking ping from %v to %s", c, ipaddr)
	out, err := d.exec(c, "ping -c 1 "+ipaddr)

	if err != nil || strings.Contains(out, "0 received, 100% packet loss") {
		logrus.Errorf("Ping from %v to %s FAILED: %q - %v", c, ipaddr, out, err)
		return fmt.Errorf("Ping failed from %v to %s: %q - %v", c, ipaddr, out, err)
	}

	logrus.Infof("Ping from %v to %s SUCCEEDED", c, ipaddr)
	return nil
}

func (d *docker) getIPAddr(c *container, dev string) (string, error) {
	out, err := d.exec(c, fmt.Sprintf("ip addr show dev %s | grep inet | head -1", dev))
	if err != nil {
		logrus.Errorf("Failed to get IP for container %q", c.containerID)
		logrus.Println(out)
	}

	parts := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(out), -1)
	if len(parts) < 2 {
		return "", fmt.Errorf("Invalid output from container %q: %s", c.containerID, out)
	}

	parts = strings.Split(parts[1], "/")
	out = strings.TrimSpace(parts[0])
	return out, err
}

func (d *docker) exec(c *container, args string) (string, error) {
	out, err := c.node.runCommand(fmt.Sprintf("docker exec %s %s", c.containerID, args))
	if err != nil {
		logrus.Println(out)
		return out, err
	}

	return out, nil
}

func (d *docker) execBG(c *container, args string) (string, error) {
	return c.node.runCommand(fmt.Sprintf("docker exec -d %s %s", c.containerID, args))
}

func (d *docker) dockerCmd(c *container, arg string) error {
	out, err := c.node.runCommand(fmt.Sprintf("docker %s %s", arg, c.containerID))
	if err != nil {
		logrus.Println(out)
		return err
	}

	return nil
}

func (d *docker) start(c *container) error {
	logrus.Infof("Starting container %s on %s", c.containerID, c.node.Name())
	return d.dockerCmd(c, "start")
}

func (d *docker) stop(c *container) error {
	logrus.Infof("Stopping container %s on %s", c.containerID, c.node.Name())
	return d.dockerCmd(c, "stop")
}

func (d *docker) rm(c *container) error {
	logrus.Infof("Removing container %s on %s", c.containerID, c.node.Name())
	d.dockerCmd(c, "kill -s 9")
	return d.dockerCmd(c, "rm -f")
}

func (d *docker) startListener(c *container, port int, protocol string) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Starting a %s listener on %v port %d", protocol, c, port)
	_, err := d.execBG(c, fmt.Sprintf("nc -lk %s -p %v -e /bin/true", protoStr, port))
	return err
}

func (d *docker) checkConnection(c *container, ipaddr, protocol string, port int) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Checking connection from %v to ip %s on port %d", c, ipaddr, port)

	_, err := d.exec(c, fmt.Sprintf("nc -z -n -v -w 1 %s %s %v", protoStr, ipaddr, port))
	if err != nil {
		logrus.Errorf("Connection from %v to ip %s on port %d FAILED", c, ipaddr, port)
	} else {
		logrus.Infof("Connection from %v to ip %s on port %d SUCCEEDED", c, ipaddr, port)
	}

	return err
}

func (d *docker) checkNoConnection(c *container, ipaddr, protocol string, port int) error {
	logrus.Infof("Expecting connection to fail from %v to %s on port %d", c, ipaddr, port)

	if err := d.checkConnection(c, ipaddr, protocol, port); err != nil {
		return nil
	}

	return fmt.Errorf("Connection SUCCEEDED on port %d from %s from %v when it should have FAILED.", port, ipaddr, c)
}

func (d *docker) cleanupDockerNetwork() error {
	logrus.Infof("Cleaning up networks on %s", d.node.Name())
	return d.node.tbnode.RunCommand("docker network ls | grep netplugin | awk '{print $2}'")
}

func (d *docker) cleanupContainers() error {
	logrus.Infof("Cleaning up containers on %s", d.node.Name())
	return d.node.tbnode.RunCommand("docker kill -s 9 `docker ps -aq`; docker rm -f `docker ps -aq`")
}

func (d *docker) startNetplugin(args string) error {
	logrus.Infof("Starting netplugin on %s", d.node.Name())
	return d.node.tbnode.RunCommandBackground("sudo " + d.node.suite.binpath + "/netplugin -plugin-mode docker -vlan-if " + d.node.suite.vlanIf + " --cluster-store " + d.node.suite.clusterStore + " " + args + "&> /tmp/netplugin.log")
}

func (d *docker) stopNetplugin() error {
	logrus.Infof("Stopping netplugin on %s", d.node.Name())
	return d.node.tbnode.RunCommand("sudo pkill netplugin")
}

func (d *docker) stopNetmaster() error {
	logrus.Infof("Stopping netmaster on %s", d.node.Name())
	return d.node.tbnode.RunCommand("sudo pkill netmaster")
}

func (d *docker) startNetmaster() error {
	logrus.Infof("Starting netmaster on %s", d.node.Name())
	dnsOpt := " --dns-enable=false "
	if d.node.suite.enableDNS {
		dnsOpt = " --dns-enable=true "
	}
	return d.node.tbnode.RunCommandBackground(d.node.suite.binpath + "/netmaster" + dnsOpt + " --cluster-store " + d.node.suite.clusterStore + " &> /tmp/netmaster.log")
}
func (d *docker) cleanupMaster() {
	logrus.Infof("Cleaning up master on %s", d.node.Name())
	vNode := d.node.tbnode
	vNode.RunCommand("etcdctl rm --recursive /contiv")
	vNode.RunCommand("etcdctl rm --recursive /contiv.io")
	vNode.RunCommand("etcdctl rm --recursive /docker")
	vNode.RunCommand("etcdctl rm --recursive /skydns")
	vNode.RunCommand("curl -X DELETE localhost:8500/v1/kv/contiv.io?recurse=true")
	vNode.RunCommand("curl -X DELETE localhost:8500/v1/kv/docker?recurse=true")
}

func (d *docker) cleanupSlave() {
	logrus.Infof("Cleaning up slave on %s", d.node.Name())
	vNode := d.node.tbnode
	vNode.RunCommand("sudo ovs-vsctl del-br contivVxlanBridge")
	vNode.RunCommand("sudo ovs-vsctl del-br contivVlanBridge")
	vNode.RunCommand("for p in `ifconfig  | grep vport | awk '{print $1}'`; do sudo ip link delete $p type veth; done")
	vNode.RunCommand("sudo rm /var/run/docker/plugins/netplugin.sock")
	vNode.RunCommand("sudo service docker restart")
}
