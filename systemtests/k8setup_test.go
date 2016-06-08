package systemtests


type kubernetes struct{
    n *node
}

(s * systemtestSuite) NewK8sExec(n *node) kubernetes {
    k8:= new(kubernetes)
    k8.node = n
    return k8
}

(k8 *kubernetes) runContainer(spec *jobSpec){

}
type container struct {
	node        *node
	containerID stringvim
	name        string
	eth0        string
}

func newContainer(node *node, containerID, name string) (*container, error) {
	cont := &container{node: node, containerID: containerID, name: name}

	out, err := cont.getIPAddr("eth0")
	if err != nil {
		return nil, err
	}
	cont.eth0 = out

	return cont, nil
}
func (k8 *kubernetes) runContainer(spec containerSpec) (*container, error) {
	var namestr, netstr, dnsStr string

	if spec.networkName != "" {
		netstr = spec.networkName

		if spec.serviceName != "" {
			netstr = spec.serviceName + "." + netstr
		}

		labels = "--labels=io.contiv.network=" + netstr
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
	logrus.Infof("Starting a Pod running %q on %s", spec.commandName, n.Name())

	cmd := fmt.Sprintf("docker run -itd %s %s %s %s %s", namestr, netstr, dnsStr, spec.imageName, spec.commandName)

	logrus.Infof("Starting container on %s with: %s", n.Name(), cmd)

	out, err := n.tbnode.RunCommandWithOutput(cmd)
	if err != nil {
		logrus.Infof("cmd %q failed: output below", cmd)
		logrus.Println(out)
		out2, err := n.tbnode.RunCommandWithOutput(fmt.Sprintf("docker logs %s", strings.TrimSpace(out)))
		if err == nil {
			logrus.Println(out2)
		} else {
			logrus.Errorf("Container id %q is invalid", strings.TrimSpace(out))
		}

		return nil, err
	}

	cont, err := newContainer(n, strings.TrimSpace(out), spec.name)
	if err != nil {
		logrus.Info(err)
		return nil, err
	}

	return cont, nil
}

func (c *kubernetes) String() string {
	return fmt.Sprintf("(container: %s (name: %q ip: %s host: %s))", c.containerID, c.name, c.eth0, c.node.Name())
}

func (c *kubernetes) checkPingFailure(ipaddr string) error {
	logrus.Infof("Expecting ping failure from %v to %s", c, ipaddr)
	if err := c.checkPing(ipaddr); err == nil {
		return fmt.Errorf("Ping succeeded when expected to fail from %v to %s", c, ipaddr)
	}

	return nil
}

func (c *kubernetes) checkPing(ipaddr string) error {
	logrus.Infof("Checking ping from %v to %s", c, ipaddr)
	out, err := c.exec("ping -c 1 " + ipaddr)

	if err != nil || strings.Contains(out, "0 received, 100% packet loss") {
		logrus.Errorf("Ping from %v to %s FAILED: %q - %v", c, ipaddr, out, err)
		return fmt.Errorf("Ping failed from %v to %s: %q - %v", c, ipaddr, out, err)
	}

	logrus.Infof("Ping from %v to %s SUCCEEDED", c, ipaddr)
	return nil
}

func (c *kubernetes) getIPAddr(dev string) (string, error) {
	out, err := c.exec(fmt.Sprintf("ip addr show dev %s | grep inet | head -1", dev))
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

func (c *kubernetes) exec(args string) (string, error) {
	out, err := c.node.runCommand(fmt.Sprintf("docker exec %s %s", c.containerID, args))
	if err != nil {
		logrus.Println(out)
		return out, err
	}

	return out, nil
}

func (c *kubernetes) execBG(args string) (string, error) {
	return c.node.runCommand(fmt.Sprintf("docker exec -d %s %s", c.containerID, args))
}

func (c *container) dockerCmd(arg string) error {
	out, err := c.node.runCommand(fmt.Sprintf("docker %s %s", arg, c.containerID))
	if err != nil {
		logrus.Println(out)
		return err
	}

	return nil
}

func (c *kubernetes) start() error {
	logrus.Infof("Starting container %s on %s", c.containerID, c.node.Name())
	return c.dockerCmd("start")
}

func (c *kubernetes) stop() error {
	logrus.Infof("Stopping container %s on %s", c.containerID, c.node.Name())
	return c.dockerCmd("stop")
}

func (c *kubernetes) rm() error {
	logrus.Infof("Removing container %s on %s", c.containerID, c.node.Name())
	c.dockerCmd("kill -s 9")
	return c.dockerCmd("rm -f")
}

func (c *kubernetes) startListener(port int, protocol string) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Starting a %s listener on %v port %d", protocol, c, port)
	_, err := c.execBG(fmt.Sprintf("nc -lk %s -p %v -e /bin/true", protoStr, port))
	return err
}

func (c *kubernetes) checkConnection(ipaddr, protocol string, port int) error {
	var protoStr string

	if protocol == "udp" {
		protoStr = "-u"
	}

	logrus.Infof("Checking connection from %v to ip %s on port %d", c, ipaddr, port)

	_, err := c.exec(fmt.Sprintf("nc -z -n -v -w 1 %s %s %v", protoStr, ipaddr, port))
	if err != nil {
		logrus.Errorf("Connection from %v to ip %s on port %d FAILED", c, ipaddr, port)
	} else {
		logrus.Infof("Connection from %v to ip %s on port %d SUCCEEDED", c, ipaddr, port)
	}

	return err
}

func (c *kubernetes) checkNoConnection(ipaddr, protocol string, port int) error {
	logrus.Infof("Expecting connection to fail from %v to %s on port %d", c, ipaddr, port)

	if err := c.checkConnection(ipaddr, protocol, port); err != nil {
		return nil
	}

	return fmt.Errorf("Connection SUCCEEDED on port %d from %s from %v when it should have FAILED.", port, ipaddr, c)
}

func (n *node) cleanupDockerNetwork() error {
	logrus.Infof("Cleaning up networks on %s", n.Name())
	return n.tbnode.RunCommand("docker network ls | grep netplugin | awk '{print $2}'")
}

func (n *node) cleanupContainers() error {
	logrus.Infof("Cleaning up containers on %s", n.Name())
	return n.tbnode.RunCommand("docker kill -s 9 `docker ps -aq`; docker rm -f `docker ps -aq`")
}
