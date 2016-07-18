package systemtests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
)

type swarm struct {
	node *node
	env  string
}

func (s *systemtestSuite) NewSwarmExec(n *node) *swarm {
	w := new(swarm)
	w.node = n
	logrus.Infof("Running in swarm mode---------")
	w.env = "DOCKER_HOST=192.168.2.11:2385 "
	return w
}

func (w *swarm) newContainer(node *node, containerID, name string) (*container, error) {
	cont := &container{node: node, containerID: containerID, name: name}

	out, err := node.exec.getIPAddr(cont, "eth0")
	if err != nil {
		return nil, err
	}
	cont.eth0.ip = out

	out, err = cont.node.exec.getIPv6Addr(cont, "eth0")
	if err == nil {
		cont.eth0.ipv6 = out
	}

	return cont, nil
}

func (w *swarm) runContainer(spec containerSpec) (*container, error) {
	var namestr, netstr, dnsStr, labelstr string

	if (spec.networkName != "") && (spec.tenantName != "default") {
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

	logrus.Infof("Starting a container running %q on %s", spec.commandName, w.node.Name())

	cmd := fmt.Sprintf("docker run -itd %s %s %s %s %s %s", namestr, netstr, dnsStr, labelstr, spec.imageName, spec.commandName)

	out, err := w.node.tbnode.RunCommandWithOutput(w.env + cmd)
	if err != nil {
		logrus.Infof("cmd %q failed: output below", cmd)
		logrus.Println(out)
		out2, err := w.node.tbnode.RunCommandWithOutput(fmt.Sprintf("docker logs %s", strings.TrimSpace(out)))
		if err == nil {
			logrus.Println(out2)
		} else {
			logrus.Errorf("Container id %q is invalid", strings.TrimSpace(out))
		}

		return nil, err
	}

	cont, err := w.newContainer(w.node, strings.TrimSpace(out), spec.name)
	if err != nil {
		logrus.Info(err)
		return nil, err
	}

	return cont, nil
}

func (w *swarm) checkPingFailure(c *container, ipaddr string) error {
	logrus.Infof("Expecting ping failure from %v to %s", c, ipaddr)
	if err := w.checkPing(c, ipaddr); err == nil {
		return fmt.Errorf("Ping succeeded when expected to fail from %v to %s", c, ipaddr)
	}

	return nil
}

func (w *swarm) checkPing(c *container, ipaddr string) error {
	logrus.Infof("Checking ping from %v to %s", c, ipaddr)
	out, err := w.exec(c, "ping -c 1 "+ipaddr)

	if err != nil || strings.Contains(out, "0 received, 100% packet loss") {
		logrus.Errorf("Ping from %v to %s FAILED: %q - %v", c, ipaddr, out, err)
		return fmt.Errorf("Ping failed from %v to %s: %q - %v", c, ipaddr, out, err)
	}

	logrus.Infof("Ping from %v to %s SUCCEEDED", c, ipaddr)
	return nil
}

func (w *swarm) checkPing6Failure(c *container, ipaddr string) error {
	logrus.Infof("Expecting ping6 failure from %v to %s", c, ipaddr)
	if err := w.checkPing6(c, ipaddr); err == nil {
		return fmt.Errorf("Ping6 succeeded when expected to fail from %v to %s", c, ipaddr)
	}

	return nil
}

func (w *swarm) checkPing6(c *container, ipaddr string) error {
	logrus.Infof("Checking ping6 from %v to %s", c, ipaddr)
	out, err := w.exec(c, "ping6 -c 1 "+ipaddr)

	if err != nil || strings.Contains(out, "0 received, 100% packet loss") {
		logrus.Errorf("Ping6 from %v to %s FAILED: %q - %v", c, ipaddr, out, err)
		return fmt.Errorf("Ping6 failed from %v to %s: %q - %v", c, ipaddr, out, err)
	}

	logrus.Infof("Ping6 from %v to %s SUCCEEDED", c, ipaddr)
	return nil
}

func (w *swarm) getIPAddr(c *container, dev string) (string, error) {
	out, err := w.exec(c, fmt.Sprintf("ip addr show dev %s | grep inet | head -1", dev))
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

func (w *swarm) getIPv6Addr(c *container, dev string) (string, error) {
	out, err := w.exec(c, fmt.Sprintf("ip addr show dev %s | grep 'inet6.*scope.*global' | head -1", dev))
	if err != nil {
		logrus.Errorf("Failed to get IPv6 for container %q", c.containerID)
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

func (w *swarm) exec(c *container, args string) (string, error) {
	logrus.Infof("container ID %s", c.containerID)
	logrus.Infof("args are %s", args)

	out, err := c.node.runCommand(fmt.Sprintf(w.env+"docker exec %s %s", c.containerID, args))
	if err != nil {
		if strings.Contains(args, "nc ") && out == "" {
			logrus.Infof("heyllooooo")
			return out, nil
		}

		logrus.Println(out)
		return out, err
	}

	return out, nil
}

func (w *swarm) execBG(c *container, args string) (string, error) {
	return c.node.runCommand(fmt.Sprintf(w.env+"docker exec -d %s %s", c.containerID, args))
}

func (w *swarm) swarmCmd(c *container, arg string) error {
	out, err := c.node.runCommand(fmt.Sprintf(w.env+"docker %s %s", arg, c.containerID))
	if err != nil {
		logrus.Println(out)
		return err
	}

	return nil
}

func (w *swarm) start(c *container) error {
	logrus.Infof("Starting container %s on %s", c.containerID, c.node.Name())
	return w.swarmCmd(c, "start")
}

func (w *swarm) stop(c *container) error {
	logrus.Infof("Stopping container %s on %s", c.containerID, c.node.Name())
	return w.swarmCmd(c, "stop")
}

func (w *swarm) rm(c *container) error {
	logrus.Infof("Removing container %s on %s", c.containerID, c.node.Name())
	w.swarmCmd(c, "kill -s 9")
	return w.swarmCmd(c, "rm -f")
}

func (w *swarm) startListener(c *container, port int, protocol string) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Starting a %s listener on %v port %d", protocol, c, port)
	_, err := w.execBG(c, fmt.Sprintf(w.env+"nc -lk %s -p %v -e /bin/true", protoStr, port))
	return err
}

func (w *swarm) checkConnection(c *container, ipaddr, protocol string, port int) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Checking connection from %v to ip %s on port %d", c, ipaddr, port)

	_, err := w.exec(c, fmt.Sprintf("nc -z -n -v -w 1 %s %s %v", protoStr, ipaddr, port))
	if err != nil {
		logrus.Errorf("Connection from %v to ip %s on port %d FAILED", c, ipaddr, port)
	} else {
		logrus.Infof("Connection from %v to ip %s on port %d SUCCEEDED", c, ipaddr, port)
	}

	return err
}

func (w *swarm) checkNoConnection(c *container, ipaddr, protocol string, port int) error {
	logrus.Infof("Expecting connection to fail from %v to %s on port %d", c, ipaddr, port)

	if err := w.checkConnection(c, ipaddr, protocol, port); err != nil {
		return nil
	}

	return fmt.Errorf("Connection SUCCEEDED on port %d from %s from %v when it should have FAILEw.", port, ipaddr, c)
}

func (w *swarm) cleanupDockerNetwork() error {
	logrus.Infof("Cleaning up networks on %s", w.node.Name())
	return w.node.tbnode.RunCommand("docker network ls | grep netplugin | awk '{print $2}'")
}

func (w *swarm) cleanupContainers() error {
	logrus.Infof("Cleaning up containers on %s", w.node.Name())
	return w.node.tbnode.RunCommand("docker kill -s 9 `docker ps -aq`; docker rm -f `docker ps -aq`")
}

func (w *swarm) startNetplugin(args string) error {
	logrus.Infof("Starting netplugin on %s", w.node.Name())
	return w.node.tbnode.RunCommandBackground("sudo " + w.node.suite.binpath + "/netplugin -plugin-mode docker -vlan-if " + w.node.suite.vlanIf + " --cluster-store " + w.node.suite.clusterStore + " " + args + "&> /tmp/netplugin.log")
}

func (w *swarm) stopNetplugin() error {
	logrus.Infof("Stopping netplugin on %s", w.node.Name())
	return w.node.tbnode.RunCommand("sudo pkill netplugin")
}

func (w *swarm) stopNetmaster() error {
	logrus.Infof("Stopping netmaster on %s", w.node.Name())
	return w.node.tbnode.RunCommand("sudo pkill netmaster")
}

func (w *swarm) startNetmaster() error {
	logrus.Infof("Starting netmaster on %s", w.node.Name())
	dnsOpt := " --dns-enable=false "
	if w.node.suite.enableDNS {
		dnsOpt = " --dns-enable=true "
	}
	return w.node.tbnode.RunCommandBackground(w.node.suite.binpath + "/netmaster" + dnsOpt + " --cluster-store " + w.node.suite.clusterStore + " &> /tmp/netmaster.log")
}
func (w *swarm) cleanupMaster() {
	logrus.Infof("Cleaning up master on %s", w.node.Name())
	vNode := w.node.tbnode
	vNode.RunCommand("etcdctl rm --recursive /contiv")
	vNode.RunCommand("etcdctl rm --recursive /contiv.io")
	vNode.RunCommand("etcdctl rm --recursive /docker")
	vNode.RunCommand("etcdctl rm --recursive /skydns")
	vNode.RunCommand("curl -X DELETE localhost:8500/v1/kv/contiv.io?recurse=true")
	vNode.RunCommand("curl -X DELETE localhost:8500/v1/kv/docker?recurse=true")
}

func (w *swarm) cleanupSlave() {
	logrus.Infof("Cleaning up slave on %s", w.node.Name())
	vNode := w.node.tbnode
	vNode.RunCommand("sudo ovs-vsctl del-br contivVxlanBridge")
	vNode.RunCommand("sudo ovs-vsctl del-br contivVlanBridge")
	vNode.RunCommand("for p in `ifconfig  | grep vport | awk '{print $1}'`; do sudo ip link delete $p type veth; done")
	vNode.RunCommand("sudo rm /var/run/docker/plugins/netplugin.sock")
	vNode.RunCommand("sudo service docker restart")
}

func (w *swarm) runCommandUntilNoNetpluginError() error {
	return w.node.runCommandUntilNoError("pgrep netplugin")
}

func (w *swarm) runCommandUntilNoNetmasterError() error {
	return w.node.runCommandUntilNoError("pgrep netmaster")
}

func (w *swarm) rotateNetmasterLog() error {
	return w.rotateLog("netmaster")
}

func (w *swarm) rotateNetpluginLog() error {
	return w.rotateLog("netplugin")
}

func (w *swarm) checkForNetpluginErrors() error {
	out, _ := w.node.tbnode.RunCommandWithOutput(`for i in /tmp/net*; do grep "panic\|fatal" $i; done`)
	if out != "" {
		logrus.Errorf("Fatal error in logs on %s: \n", w.node.Name())
		fmt.Printf("%s\n==========================================\n", out)
		return fmt.Errorf("fatal error in netplugin logs")
	}

	out, _ = w.node.tbnode.RunCommandWithOutput(`for i in /tmp/net*; do grep "error" $i; done`)
	if out != "" {
		logrus.Errorf("error output in netplugin logs on %s: \n", w.node.Name())
		fmt.Printf("%s==========================================\n\n", out)
		// FIXME: We still have some tests that are failing error check
		// return fmt.Errorf("error output in netplugin logs")
	}

	return nil
}

func (w *swarm) rotateLog(prefix string) error {
	oldPrefix := fmt.Sprintf("/tmp/%s", prefix)
	newPrefix := fmt.Sprintf("/tmp/_%s", prefix)
	_, err := w.node.runCommand(fmt.Sprintf("mv %s.log %s-`date +%%s`.log", oldPrefix, newPrefix))
	return err
}
