package systemtests

import (
	//"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	//	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
	//	"os"
	//	"strconv"
	//	"strings"
	//"time"
)

func (s *systemtestSuite) TestACIMode(c *C) {
	if s.fwdMode == "routing" || s.basicInfo.Scheduler == "k8" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
		FwdMode:          "bridge",
	}), IsNil)
	c.Assert(s.cli.NetworkPost(&client.Network{
		TenantName:  "default",
		NetworkName: s.infoGlob.Network,
		Subnet:      s.infoGlob.Subnet,
		Gateway:     s.infoGlob.Gateway,
		Encap:       s.infoGlob.Encap,
	}), IsNil)

	err := s.nodes[0].checkSchedulerNetworkCreated(s.infoGlob.Network, false)
	c.Assert(err, IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  "default",
		NetworkName: s.infoGlob.Network,
		GroupName:   "epga",
	}), IsNil)

	err = s.nodes[0].exec.checkSchedulerNetworkCreated("epga", true)
	c.Assert(err, IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  "default",
		NetworkName: s.infoGlob.Network,
		GroupName:   "epgb",
	}), IsNil)

	err = s.nodes[0].checkSchedulerNetworkCreated("epgb", true)
	c.Assert(err, IsNil)

	containersA, err := s.runContainersOnNode(s.basicInfo.Containers, s.infoGlob.Network, "", "epga", s.nodes[0])
	c.Assert(err, IsNil)
	containersB, err := s.runContainersOnNode(s.basicInfo.Containers, s.infoGlob.Network, "", "epgb", s.nodes[0])
	c.Assert(err, IsNil)

	// Verify cA1 can ping cA2
	c.Assert(s.pingTest(containersA), IsNil)
	// Verify cB1 can ping cB2
	c.Assert(s.pingTest(containersB), IsNil)
	// Verify cA1 cannot ping cB1
	c.Assert(s.pingFailureTest(containersA, containersB), IsNil)

	log.Infof("Triggering netplugin restart")
	node1 := s.nodes[0]
	c.Assert(node1.stopNetplugin(), IsNil)
	c.Assert(node1.rotateLog("netplugin"), IsNil)
	c.Assert(node1.startNetplugin(""), IsNil)
	c.Assert(node1.runCommandUntilNoError("pgrep netplugin"), IsNil)
	time.Sleep(20 * time.Second)

	// Verify cA1 can ping cA2
	c.Assert(s.pingTest(containersA), IsNil)
	// Verify cB1 can ping cB2
	c.Assert(s.pingTest(containersB), IsNil)
	// Verify cA1 cannot ping cB1
	c.Assert(s.pingFailureTest(containersA, containersB), IsNil)

	c.Assert(s.removeContainers(containersA), IsNil)
	c.Assert(s.removeContainers(containersB), IsNil)
	c.Assert(s.cli.EndpointGroupDelete("default", "epga"), IsNil)
	c.Assert(s.cli.EndpointGroupDelete("default", "epgb"), IsNil)
	c.Assert(s.cli.NetworkDelete("default", s.infoGlob.Network), IsNil)
}

func (s *systemtestSuite) TestACIPingGateway(c *C) {
	if s.fwdMode == "routing" || s.basicInfo.Scheduler == "k8" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
		FwdMode:          "bridge",
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.infoGlob.Tenant,
	}), IsNil)

	containersA := []*container{}

	//for i := 0; i < s.basicInfo.Iterations; i++ {

	c.Assert(s.cli.NetworkPost(&client.Network{
		TenantName:  s.infoGlob.Tenant,
		NetworkName: s.infoGlob.Network,
		Subnet:      s.infoGlob.Subnet,
		Gateway:     s.infoGlob.Gateway,
		Encap:       s.infoGlob.Encap,
	}), IsNil)

	err := s.nodes[0].checkSchedulerNetworkCreated(s.infoGlob.Network, false)
	c.Assert(err, IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.infoGlob.Tenant,
		NetworkName: s.infoGlob.Network,
		GroupName:   "epga",
	}), IsNil)

	err = s.nodes[0].exec.checkSchedulerNetworkCreated("epga", true)
	c.Assert(err, IsNil)

	for i := 0; i < s.basicInfo.Iterations; i++ {
		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.infoGlob.Tenant,
			EndpointGroups: []string{"epga"},
			AppProfileName: "profile1",
		}), IsNil)
		time.Sleep(5 * time.Second)

		//s.basicInfo.Iterations = 2
		//for i := 0; i < s.basicInfo.Iterations; i++ {
		containers, err := s.runContainersInGroups(s.basicInfo.Containers, s.infoGlob.Network, s.infoGlob.Tenant, []string{"epga"})
		c.Assert(err, IsNil)

		for key := range containers {
			containersA = append(containersA, key)
		}

		// Verify containers in A can ping default gateway
		c.Assert(s.pingTestByName(containersA, s.infoGlob.Gateway), IsNil)

		c.Assert(s.removeContainers(containersA), IsNil)
		containersA = nil
		//}
		c.Assert(s.cli.AppProfileDelete(s.infoGlob.Tenant, "profile1"), IsNil)
		time.Sleep(time.Second * 5)
	}
	c.Assert(s.cli.EndpointGroupDelete(s.infoGlob.Tenant, "epga"), IsNil)
	c.Assert(s.cli.NetworkDelete(s.infoGlob.Tenant, s.infoGlob.Network), IsNil)
	//}
}

func (s *systemtestSuite) TestACIProfile(c *C) {
	if s.fwdMode == "routing" || s.basicInfo.Scheduler == "k8" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
		FwdMode:          "bridge",
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.infoGlob.Tenant,
	}), IsNil)

	containersA := []*container{}
	containersB := []*container{}

	for i := 0; i < 1; i++ {
		log.Infof(">>ITERATION #%d<<", i)
		c.Assert(s.cli.NetworkPost(&client.Network{
			TenantName:  s.infoGlob.Tenant,
			NetworkName: s.infoGlob.Network,
			Subnet:      s.infoGlob.Subnet,
			Gateway:     s.infoGlob.Gateway,
			Encap:       s.infoGlob.Encap,
		}), IsNil)

		err := s.nodes[0].checkSchedulerNetworkCreated(s.infoGlob.Network, false)
		c.Assert(err, IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.infoGlob.Tenant,
			NetworkName: s.infoGlob.Network,
			GroupName:   "epgA",
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.infoGlob.Tenant,
			NetworkName: s.infoGlob.Network,
			GroupName:   "epgB",
		}), IsNil)

		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.infoGlob.Tenant,
			EndpointGroups: []string{"epgA", "epgB"},
			AppProfileName: "profile1",
		}), IsNil)

		time.Sleep(5 * time.Second)
		//cA1, err := s.nodes[0].runContainer(containerSpec{networkName: "epgA/aciTenant"})
		//c.Assert(err, IsNil)

		// Verify cA1 can ping default gateway
		//c.Assert(cA1.checkPingWithCount("20.1.1.254", 5), IsNil)

		//cB1, err := s.nodes[0].runContainer(containerSpec{networkName: "epgB/aciTenant"})
		//c.Assert(err, IsNil)

		groups := []string{"epgA", "epgB"}
		containers, err := s.runContainersInGroups(s.basicInfo.Containers, s.infoGlob.Network, s.infoGlob.Tenant, groups)
		c.Assert(err, IsNil)
		time.Sleep(time.Second * 20)
		for key, value := range containers {
			if value == "epgA/aciTenant" {
				containersA = append(containersA, key)
			} else {
				containersB = append(containersB, key)
			}
		}
		log.Infof("-----contA %s", containersA)
		log.Infof("-----contB %s", containersB)
		/*
			// Verify cA1 cannot ping cB1
			c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)
			// Verify cB1 can ping default gateway
			c.Assert(cB1.checkPingWithCount("20.1.1.254", 5), IsNil)

			// Create a policy that allows ICMP and apply between A and B
			c.Assert(s.cli.PolicyPost(&client.Policy{
				PolicyName: "policyAB",
				TenantName: "aciTenant",
			}), IsNil)

			c.Assert(s.cli.RulePost(&client.Rule{
				RuleID:            "1",
				PolicyName:        "policyAB",
				TenantName:        "aciTenant",
				FromEndpointGroup: "epgA",
				Direction:         "in",
				Protocol:          "icmp",
				Action:            "allow",
			}), IsNil)

			c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
				TenantName:  "aciTenant",
				NetworkName: "aciNet",
				Policies:    []string{"policyAB"},
				GroupName:   "epgB",
			}), IsNil)
			time.Sleep(time.Second * 5)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgB",
				cB1), IsNil)

			// Verify cA1 can ping cB1
			cA1.checkPingWithCount(cB1.eth0.ip, 5)
			cA1.checkPingWithCount(cB1.eth0.ip, 5)
			cA1.checkPingWithCount(cB1.eth0.ip, 5)
			c.Assert(cA1.checkPingWithCount(cB1.eth0.ip, 5), IsNil)

			// Verify TCP is not allowed.
			containers := []*container{cA1, cB1}
			from := []*container{cA1}
			to := []*container{cB1}

			c.Assert(s.startListeners(containers, []int{8000, 8001}), IsNil)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)

			// Add a rule to allow 8000
			c.Assert(s.cli.RulePost(&client.Rule{
				RuleID:            "2",
				PolicyName:        "policyAB",
				TenantName:        "aciTenant",
				FromEndpointGroup: "epgA",
				Direction:         "in",
				Protocol:          "tcp",
				Port:              8000,
				Action:            "allow",
			}), IsNil)
			time.Sleep(time.Second * 5)
			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgB",
				cB1), IsNil)

			c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA1.checkPingWithCount(cB1.eth0.ip, 5), IsNil)

			// Add a rule to allow 8001
			c.Assert(s.cli.RulePost(&client.Rule{
				RuleID:            "3",
				PolicyName:        "policyAB",
				TenantName:        "aciTenant",
				FromEndpointGroup: "epgA",
				Direction:         "in",
				Protocol:          "tcp",
				Port:              8001,
				Action:            "allow",
			}), IsNil)
			//cA1.checkPing("20.1.1.254")
			//cB1.checkPing("20.1.1.254")
			time.Sleep(time.Second * 5)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgB",
				cB1), IsNil)

			c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA1.checkPingWithCount(cB1.eth0.ip, 5), IsNil)

			// Delete ICMP rule
			c.Assert(s.cli.RuleDelete("aciTenant", "policyAB", "1"), IsNil)
			time.Sleep(time.Second * 5)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgB",
				cB1), IsNil)

			c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)

			// Delete TCP 8000 rule
			c.Assert(s.cli.RuleDelete("aciTenant", "policyAB", "2"), IsNil)
			time.Sleep(time.Second * 5)
			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile1",
				"epgB",
				cB1), IsNil)

			c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)

			// Delete the app profile
			c.Assert(s.cli.AppProfileDelete("aciTenant", "profile1"), IsNil)
			time.Sleep(time.Second * 5)
			//cA1.checkPingWithCount("20.1.1.254", 5)
			//cB1.checkPingWithCount("20.1.1.254", 5)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)

			// Create the app profile with a different name
			c.Assert(s.cli.AppProfilePost(&client.AppProfile{
				TenantName:     "aciTenant",
				EndpointGroups: []string{"epgA", "epgB"},
				AppProfileName: "profile2",
			}), IsNil)
			time.Sleep(time.Second * 5)
			c.Assert(checkACILearning("aciTenant",
				"profile2",
				"epgA",
				cA1), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile2",
				"epgB",
				cB1), IsNil)

			//cA1.checkPingWithCount("20.1.1.254", 5)
			//cB1.checkPingWithCount("20.1.1.254", 5)
			cA2, err := s.nodes[0].runContainer(containerSpec{networkName: "epgA/aciTenant"})
			c.Assert(err, IsNil)
			cB2, err := s.nodes[0].runContainer(containerSpec{networkName: "epgB/aciTenant"})
			c.Assert(err, IsNil)
			time.Sleep(time.Second * 10)
			from = []*container{cA2}
			to = []*container{cB2}
			c.Assert(s.startListeners([]*container{cA2, cB2}, []int{8000, 8001}), IsNil)

			c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA2.checkPingFailureWithCount(cB2.eth0.ip, 5), IsNil)

			// Add TCP 8000 rule.
			c.Assert(s.cli.RulePost(&client.Rule{
				RuleID:            "2",
				PolicyName:        "policyAB",
				TenantName:        "aciTenant",
				FromEndpointGroup: "epgA",
				Direction:         "in",
				Protocol:          "tcp",
				Port:              8000,
				Action:            "allow",
			}), IsNil)
			err = errors.New("Forced")
			//c.Assert(err, IsNil)
			time.Sleep(time.Second * 5)
			c.Assert(checkACILearning("aciTenant",
				"profile2",
				"epgA",
				cA2), IsNil)

			c.Assert(checkACILearning("aciTenant",
				"profile2",
				"epgB",
				cB2), IsNil)

			//cA1.checkPingWithCount("20.1.1.254", 5)
			//cB1.checkPingWithCount("20.1.1.254", 5)

			c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA2.checkPingFailureWithCount(cB2.eth0.ip, 5), IsNil)

			// Delete the app profile
			c.Assert(s.cli.AppProfileDelete("aciTenant", "profile2"), IsNil)
			time.Sleep(time.Second * 5)
			//cA1.checkPingWithCount("20.1.1.254", 5)
			//cB1.checkPingWithCount("20.1.1.254", 5)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
			c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
			c.Assert(cA2.checkPingFailureWithCount(cB2.eth0.ip, 5), IsNil)

			c.Assert(s.removeContainers([]*container{cA1, cB1, cA2, cB2}), IsNil)
			c.Assert(s.cli.EndpointGroupDelete("aciTenant", "epgA"), IsNil)
			c.Assert(s.cli.EndpointGroupDelete("aciTenant", "epgB"), IsNil)
			c.Assert(s.cli.RuleDelete("aciTenant", "policyAB", "2"), IsNil)
			c.Assert(s.cli.RuleDelete("aciTenant", "policyAB", "3"), IsNil)
			c.Assert(s.cli.PolicyDelete("aciTenant", "policyAB"), IsNil)
		*/
		c.Assert(s.removeContainers(append(containersA, containersB...)), IsNil)
		c.Assert(s.cli.EndpointGroupDelete(s.infoGlob.Tenant, "epgA"), IsNil)
		c.Assert(s.cli.EndpointGroupDelete(s.infoGlob.Tenant, "epgB"), IsNil)
		c.Assert(s.cli.NetworkDelete(s.infoGlob.Tenant, s.infoGlob.Network), IsNil)
	}
}
