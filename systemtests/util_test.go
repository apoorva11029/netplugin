package systemtests

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func (s *systemtestSuite) checkConnectionPair(containers1, containers2 []*container, port int) error {
	for _, cont := range containers1 {
		for _, cont2 := range containers2 {
			if err := cont.node.exec.checkConnection(cont, cont2.eth0.ip, "tcp", port); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *systemtestSuite) runContainersInGroups(num int, netName string, tenantName string, groupNames []string) (map[*container]string, error) {
	containers := map[*container]string{}
	for _, groupName := range groupNames {
		names := []string{}

		for i := 0; i < num; i++ {
			names = append(names, fmt.Sprintf("grp-%s-%d", groupName, i))
		}

		// XXX we don't use anything but private for this function right now
		conts, err := s.runContainersInService(num, groupName, netName, "", names)
		if err != nil {
			return nil, err
		}

		for _, cont := range conts {
			containers[cont] = groupName
		}
	}

	return containers, nil
}

func (s *systemtestSuite) runContainersInService(num int, serviceName, networkName string, tenantName string, names []string) ([]*container, error) {
	containers := []*container{}
	mutex := sync.Mutex{}

	if networkName == "" {
		networkName = "private"
	}

	errChan := make(chan error)

	for i := 0; i < num; i++ {
		go func(i int) {
			nodeNum := i % len(s.nodes)
			var name string

			mutex.Lock()
			if len(names) > 0 {
				name = names[0]
				names = names[1:]
			}
			mutex.Unlock()

			if name == "" {
				name = fmt.Sprintf("%s-srv%d-%d", strings.Replace(networkName, "/", "-", -1), i, nodeNum)
			}

			spec := containerSpec{
				imageName:   "alpine",
				networkName: networkName,
				name:        name,
				serviceName: serviceName,
				tenantName:  tenantName,
			}

			cont, err := s.nodes[nodeNum].exec.runContainer(spec)
			if err != nil {
				errChan <- err
			}

			mutex.Lock()
			containers = append(containers, cont)
			mutex.Unlock()

			errChan <- nil
		}(i)
	}

	for i := 0; i < num; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return containers, nil
}

func (s *systemtestSuite) runContainers(num int, withService bool, networkName string, tenantName string,
	names []string, labels []string) ([]*container, error) {
	containers := []*container{}
	mutex := sync.Mutex{}

	if networkName == "" {
		networkName = "private"
	}

	errChan := make(chan error)

	for i := 0; i < num; i++ {
		go func(i int) {
			nodeNum := i % len(s.nodes)
			var name string

			mutex.Lock()
			if len(names) > 0 {
				name = names[0]
				names = names[1:]
			}
			mutex.Unlock()

			if name == "" {
				name = fmt.Sprintf("%s-srv%d-%d", strings.Replace(networkName, "/", "-", -1), i, nodeNum)
			}

			var serviceName string

			if withService {
				serviceName = name
			}

			spec := containerSpec{
				imageName:   "alpine",
				networkName: networkName,
				name:        name,
				serviceName: serviceName,
				tenantName:  tenantName,
			}
			if len(labels) > 0 {
				spec.labels = append(spec.labels, labels...)
			}

			cont, err := s.nodes[nodeNum].exec.runContainer(spec)
			if err != nil {
				errChan <- err
			}

			mutex.Lock()
			containers = append(containers, cont)
			mutex.Unlock()

			errChan <- nil
		}(i)
	}

	for i := 0; i < num; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return containers, nil
}

func (s *systemtestSuite) runContainersSerial(num int, withService bool, networkName string, tenantName string, names []string) ([]*container, error) {
	containers := []*container{}
	mutex := sync.Mutex{}

	if networkName == "" {
		networkName = "private"
	}

	for i := 0; i < num; i++ {
		nodeNum := i % len(s.nodes)
		var name string

		mutex.Lock()
		if len(names) > 0 {
			name = names[0]
			names = names[1:]
		}
		mutex.Unlock()

		if name == "" {
			name = fmt.Sprintf("%s-srv%d-%d", strings.Replace(networkName, "/", "-", -1), i, nodeNum)
		}

		var serviceName string

		if withService {
			serviceName = name
		}

		spec := containerSpec{
			imageName:   "alpine",
			networkName: networkName,
			name:        name,
			serviceName: serviceName,
			tenantName:  tenantName,
		}

		cont, err := s.nodes[nodeNum].exec.runContainer(spec)
		if err != nil {
			return nil, err
		}

		mutex.Lock()
		containers = append(containers, cont)
		mutex.Unlock()

	}

	return containers, nil
}

func (s *systemtestSuite) runContainersWithDNS(num int, tenantName, networkName, serviceName string) ([]*container, error) {
	containers := []*container{}
	mutex := sync.Mutex{}

	errChan := make(chan error)

	// Get the dns server for the network
	dnsServer, err := s.getNetworkDNSServer(tenantName, networkName)
	if err != nil {
		logrus.Errorf("Error getting DNS server for network %s/%s", networkName, tenantName)
		return nil, err
	}

	docknetName := fmt.Sprintf("%s/%s", networkName, tenantName)
	if tenantName == "default" {
		docknetName = networkName
	}
	docSrvName := docknetName
	if serviceName != "" {
		docSrvName = fmt.Sprintf("%s.%s", serviceName, docknetName)
	}

	for i := 0; i < num; i++ {
		go func(i int) {
			nodeNum := i % len(s.nodes)
			name := fmt.Sprintf("%s-srv%d-%d", strings.Replace(docSrvName, "/", "-", -1), i, nodeNum)

			spec := containerSpec{
				imageName:   "alpine",
				networkName: networkName,
				name:        name,
				serviceName: serviceName,
				dnsServer:   dnsServer,
				tenantName:  tenantName,
			}
			logrus.Infof("will run containers now")
			cont, err := s.nodes[nodeNum].exec.runContainer(spec)
			logrus.Infof("done tunnint containers ")
			if err != nil {
				errChan <- err
			}

			mutex.Lock()
			containers = append(containers, cont)
			mutex.Unlock()

			errChan <- nil
		}(i)
	}

	for i := 0; i < num; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return containers, nil
}

func (s *systemtestSuite) pingTest(containers []*container) error {
	ips := []string{}
	v6ips := []string{}
	for _, cont := range containers {
		ips = append(ips, cont.eth0.ip)
		if cont.eth0.ipv6 != "" {
			v6ips = append(v6ips, cont.eth0.ipv6)
		}
	}

	errChan := make(chan error, len(containers)*len(ips))

	for _, cont := range containers {
		for _, ip := range ips {
			go func(cont *container, ip string) { errChan <- cont.node.exec.checkPing(cont, ip) }(cont, ip)
		}
	}

	for i := 0; i < len(containers)*len(ips); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	if len(v6ips) > 0 {
		v6errChan := make(chan error, len(containers)*len(v6ips))

		for _, cont := range containers {
			for _, ipv6 := range v6ips {
				go func(cont *container, ipv6 string) { v6errChan <- cont.node.exec.checkPing6(cont, ipv6) }(cont, ipv6)
			}
		}

		for i := 0; i < len(containers)*len(v6ips); i++ {
			if err := <-v6errChan; err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *systemtestSuite) pingTestByName(containers []*container, hostName string) error {

	errChan := make(chan error, len(containers))

	for _, cont := range containers {
		go func(cont *container, hostName string) { errChan <- cont.node.exec.checkPing(cont, hostName) }(cont, hostName)
	}

	for i := 0; i < len(containers); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) pingFailureTest(containers1 []*container, containers2 []*container) error {
	errChan := make(chan error, len(containers1)*len(containers2))

	for _, cont1 := range containers1 {
		for _, cont2 := range containers2 {
			go func(cont1 *container, cont2 *container) {
				errChan <- cont1.node.exec.checkPingFailure(cont1, cont2.eth0.ip)
			}(cont1, cont2)
		}
	}

	for i := 0; i < len(containers1)*len(containers2); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) removeContainers(containers []*container) error {
	errChan := make(chan error, len(containers))
	for _, cont := range containers {
		go func(cont *container) { errChan <- cont.node.exec.rm(cont) }(cont)
	}

	for range containers {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) startListeners(containers []*container, ports []int) error {
	errChan := make(chan error, len(containers)*len(ports))

	for _, cont := range containers {
		for _, port := range ports {
			go func(cont *container, port int) { errChan <- cont.node.exec.startListener(cont, port, "tcp") }(cont, port)
		}
	}

	for i := 0; i < len(containers)*len(ports); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) checkConnections(containers []*container, port int) error {
	ips := []string{}
	for _, cont := range containers {
		ips = append(ips, cont.eth0.ip)
	}

	endChan := make(chan error, len(containers))

	for _, cont := range containers {
		for _, ip := range ips {
			if cont.eth0.ip == ip {
				continue
			}

			go func(cont *container, ip string, port int) {
				endChan <- cont.node.exec.checkConnection(cont, ip, "tcp", port)
			}(cont, ip, port)
		}
	}

	for i := 0; i < len(containers)*(len(ips)-1); i++ {
		if err := <-endChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) checkNoConnections(containers []*container, port int) error {
	ips := []string{}
	for _, cont := range containers {
		ips = append(ips, cont.eth0.ip)
	}

	endChan := make(chan error, len(containers))

	for _, cont := range containers {
		for _, ip := range ips {
			if cont.eth0.ip == ip {
				continue
			}

			go func(cont *container, ip string, port int) {
				endChan <- cont.node.exec.checkNoConnection(cont, ip, "tcp", port)
			}(cont, ip, port)
		}
	}

	for i := 0; i < len(containers)*(len(ips)-1); i++ {
		if err := <-endChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) checkConnectionsAcrossGroup(containers map[*container]string, port int, expFail bool) error {
	groups := map[string][]*container{}

	for cont1, group := range containers {
		if _, ok := groups[group]; !ok {
			groups[group] = []*container{}
		}

		groups[group] = append(groups[group], cont1)
	}

	for cont1, group := range containers {
		for group2, conts := range groups {
			if group != group2 {
				for _, cont := range conts {
					err := cont1.node.exec.checkConnection(cont1, cont.eth0.ip, "tcp", port)
					if !expFail && err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *systemtestSuite) checkConnectionsWithinGroup(containers map[*container]string, port int) error {
	groups := map[string][]*container{}

	for cont1, group := range containers {
		if _, ok := groups[group]; !ok {
			groups[group] = []*container{}
		}

		groups[group] = append(groups[group], cont1)
	}

	for cont1, group := range containers {
		for group2, conts := range groups {
			if group == group2 {
				for _, cont := range conts {
					if err := cont1.node.exec.checkConnection(cont1, cont.eth0.ip, "tcp", port); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *systemtestSuite) checkPingContainersInNetworks(containers map[*container]string) error {
	networks := map[string][]*container{}

	for cont1, network := range containers {
		if _, ok := networks[network]; !ok {
			networks[network] = []*container{}
		}

		networks[network] = append(networks[network], cont1)
	}

	for cont1, network := range containers {
		for network2, conts := range networks {
			if network2 == network {
				for _, cont := range conts {
					if err := cont1.node.exec.checkPing(cont1, cont.eth0.ip); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *systemtestSuite) checkAllConnection(netContainers map[*container]string, groupContainers map[*container]string) error {
	if err := s.checkPingContainersInNetworks(netContainers); err != nil {
		return err
	}

	if err := s.checkConnectionsWithinGroup(groupContainers, 8000); err != nil {
		return err
	}

	if err := s.checkConnectionsWithinGroup(groupContainers, 8001); err != nil {
		return err
	}

	if err := s.checkConnectionsAcrossGroup(groupContainers, 8000, false); err != nil {
		return err
	}

	if err := s.checkConnectionsAcrossGroup(groupContainers, 8001, true); err != nil {
		return fmt.Errorf("Connections across group achieved for port 8001")
	}

	return nil
}

func (s *systemtestSuite) pingFailureTestDifferentNode(containers1 []*container, containers2 []*container) error {

	count := 0

	for _, cont1 := range containers1 {
		for _, cont2 := range containers2 {
			if cont1.node != cont2.node {
				count++
			}
		}
	}
	errChan := make(chan error, count)

	for _, cont1 := range containers1 {
		for _, cont2 := range containers2 {
			if cont1.node != cont2.node {
				go func(cont1 *container, cont2 *container) {
					errChan <- cont1.node.exec.checkPingFailure(cont1, cont2.eth0.ip)
				}(cont1, cont2)
			}
		}
	}

	for i := 0; i < count; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) pingTestToNonContainer(containers []*container, nonContIps []string) error {

	errChan := make(chan error, len(containers)*len(nonContIps))

	for _, cont := range containers {
		for _, ip := range nonContIps {
			go func(cont *container, ip string) { errChan <- cont.node.exec.checkPing(cont, ip) }(cont, ip)
		}
	}

	for i := 0; i < len(containers)*len(nonContIps); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

func (s *systemtestSuite) getJSON(url string, target interface{}) error {
	content, err := s.nodes[0].runCommand(fmt.Sprintf("curl -s %s", url))
	if err != nil {
		logrus.Errorf("Error getting curl output: Err %v", err)
		return err
	}
	return json.Unmarshal([]byte(content), target)
}

func (s *systemtestSuite) clusterStoreGet(path string) (string, error) {
	if strings.Contains(s.basicInfo.ClusterStore, "etcd://") {
		var etcdKv map[string]interface{}

		// Get from etcd
		etcdURL := strings.Replace(s.basicInfo.ClusterStore, "etcd://", "http://", 1)
		etcdURL = etcdURL + "/v2/keys" + path

		// get json from etcd
		err := s.getJSON(etcdURL, &etcdKv)
		if err != nil {
			logrus.Errorf("Error getting json from host. Err: %v", err)
			return "", err
		}

		node, ok := etcdKv["node"]
		if !ok {
			logrus.Errorf("Invalid json from etcd. %+v", etcdKv)
			return "", errors.New("node not found")
		}
		value, ok := node.(map[string]interface{})["value"]
		if !ok {
			logrus.Errorf("Invalid json from etcd. %+v", etcdKv)
			return "", errors.New("Value not found")
		}

		return value.(string), nil
	} else if strings.Contains(s.basicInfo.ClusterStore, "consul://") {
		var consulKv []map[string]interface{}

		// Get from consul
		consulURL := strings.Replace(s.basicInfo.ClusterStore, "consul://", "http://", 1)
		consulURL = consulURL + "/v1/kv" + path

		// get kv json from consul
		err := s.getJSON(consulURL, &consulKv)
		if err != nil {
			return "", err
		}

		value, ok := consulKv[0]["Value"]
		if !ok {
			logrus.Errorf("Invalid json from consul. %+v", consulKv)
			return "", errors.New("Value not found")
		}

		retVal, err := base64.StdEncoding.DecodeString(value.(string))
		return string(retVal), err
	} else {
		// Unknown cluster store
		return "", errors.New("Unknown cluster store")
	}
}

func (s *systemtestSuite) getNetworkStates() ([]map[string]interface{}, error) {
	var networkList []map[string]interface{}

	err := s.getJSON("localhost:9999/networks", &networkList)
	if err != nil {
		logrus.Errorf("Error getting json from host. Err: %v", err)
		return nil, err
	}

	return networkList, err
}

func (s *systemtestSuite) getNetworkDNSServer(tenant, network string) (string, error) {
	netList, err := s.getNetworkStates()
	if err != nil {
		return "", err
	}
	logrus.Infof("%s----------", netList)
	logrus.Infof("tenant name is %s", tenant)
	logrus.Infof("netwirk name is %s", network)
	for _, net := range netList {
		if net["tenant"].(string) == tenant && net["networkName"].(string) == network {
			dnsServer := net["dnsServer"].(string)

			if dnsServer == "" {
				logrus.Infof("Network %s/%s does not have a dns server", network, tenant)
				return "", errors.New("No DNS server in network")
			}
			logrus.Infof("Gor dns server %s for network %s/%s", dnsServer, network, tenant)
			return dnsServer, nil
		}
	}

	return "", errors.New("Network not found")
}
func (s *systemtestSuite) checkConnectionToService(containers []*container, ips []string, port int, protocol string) error {

	for _, cont := range containers {
		for _, ip := range ips {
			if err := cont.node.exec.checkConnection(cont, ip, "tcp", port); err != nil {
				return err
			}
		}
	}
	return nil
}

//ports is of the form 80:8080:TCP
func (s *systemtestSuite) startListenersOnProviders(containers []*container, ports []string) error {

	portList := []int{}

	for _, port := range ports {
		p := strings.Split(port, ":")
		providerPort, _ := strconv.Atoi(p[1])
		portList = append(portList, providerPort)
	}

	errChan := make(chan error, len(containers)*len(portList))

	for _, cont := range containers {
		for _, port := range portList {
			go func(cont *container, port int) { errChan <- cont.node.exec.startListener(cont, port, "tcp") }(cont, port)
		}
	}

	for i := 0; i < len(containers)*len(portList); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

func (s *systemtestSuite) runContainersOnNode(num int, networkName string, n *node) ([]*container, error) {
	containers := []*container{}
	mutex := sync.Mutex{}

	errChan := make(chan error)

	for i := 0; i < num; i++ {
		go func(i int) {
			spec := containerSpec{
				imageName:   "alpine",
				networkName: networkName,
				name:        fmt.Sprintf("%s-%d-%s", n.Name(), i, randSeq(16)),
			}

			cont, err := n.exec.runContainer(spec)
			if err != nil {
				errChan <- err
			}

			mutex.Lock()
			containers = append(containers, cont)
			mutex.Unlock()

			errChan <- nil
		}(i)
	}

	for i := 0; i < num; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return containers, nil
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (c *container) String() string {
	return fmt.Sprintf("(container: %s (name: %q ip: %s ipv6: %s host: %s))", c.containerID, c.name, c.eth0.ip, c.eth0.ipv6, c.node.Name())
}

//Helper functions for JSON file handling
func toJSON(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return string(bytes)
}

func (p InfoGlob) toString() string {
	return toJSON(p)
}

func (p InfoHost) toString() string {
	return toJSON(p)
}

func (p BasicInfo) toString() string {
	return toJSON(p)
}

//Function to extract ACI Info from JSON files
func getInfo(file string) ([]BasicInfo, []InfoHost, []InfoGlob) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var b []BasicInfo
	json.Unmarshal(raw, &b)
	var c []InfoHost
	json.Unmarshal(raw, &c)
	var d []InfoGlob
	json.Unmarshal(raw, &d)

	return b, c, d
}

//Function to get the master node for ACI mode
func getMaster(file string) (BasicInfo, InfoHost, InfoGlob) {
	infosbasic, infoshost, infosglob := getInfo(file)

	var mastbasic BasicInfo
	for _, p := range infosbasic {
		if p.Master == true {
			mastbasic = p
			break
		}
	}

	var masthost InfoHost
	for _, p := range infoshost {
		if p.Master == true {
			masthost = p
			break
		}
	}

	var mastglob InfoGlob
	for _, p := range infosglob {
		if p.Master == true {
			mastglob = p
			break
		}
	}
	return mastbasic, masthost, mastglob
}

func (s *systemtestSuite) SetUpSuiteBaremetal(c *C) {

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
  s.CheckNetDemoInstallation(c)
	logrus.Info("Pulling alpine on all nodes")

	s.baremetal.IterateNodes(func(node vagrantssh.TestbedNode) error {
		//fmt.Printf("\n\t Node Name is %s", node.nodeName)
		node.RunCommand("sudo rm /tmp/*net*")
		node.RunCommand("touch /home/admin/GAURAV.txt")
		return node.RunCommand("docker pull alpine")
	})

	//Copying binaries
	s.copyBinary("netmaster")
	s.copyBinary("netplugin")
	s.copyBinary("netctl")
	s.copyBinary("contivk8s")
}

func (s *systemtestSuite) SetUpSuiteVagrant(c *C) {
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
}

func (s *systemtestSuite) BaremetalSetup() {
	cmd := exec.Command("wget", "https://raw.githubusercontent.com/contiv/demo/master/net/net_demo_installer")
	cmd.Run()
	os.Chmod("net_demo_installer", 0777)
	if s.basicInfo.AciMode == "on" {
		cmd = exec.Command("./net_demo_installer", "-ars")
	} else {
		cmd = exec.Command("./net_demo_installer", "-rs")
	}
	// setup log file
	file, err := os.Create("server.log")
	if err != nil {
		logrus.Infof("no err here")
	}
	cmd.Stdout = file
	cmd.Run()
	logrus.Infof("Done running net demo ------------------")
}

//Function to check if net_demo_installer script ran properly.
//Uses the output of docker info on all the nodes in the swarm cluster.
func (s *systemtestSuite) CheckNetDemoInstallation(c *C) {
	outChan := make(chan string, 100)
	mystr := "docker info | grep Nodes"
  var err, out, out1 string
	out1, _ = s.nodes[0].runCommand(mystr)
	if out1 == "" {
		err = "The script net_demo_installer didn't run properly."
		c.Assert(err, Equals, "")
	}
	for i := 1; i < len(s.nodes); i++ {
			out, _ = s.nodes[i].runCommand(mystr)
			outChan <- out
			if out != out1 {
				err = "The script net_demo_installer didn't run properly."
				break
			}
	}
	logrus.Infof("Deleting files related to net_demo_installer")
	os.Remove("genInventoryFile.py")
	os.RemoveAll("./.gen")
	os.RemoveAll("./ansible")
	os.Remove("server.log")
	os.Remove("net_demo_installer")
	c.Assert(err, Equals, "")
}
