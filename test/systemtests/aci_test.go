package systemtests

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	//	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
	//	"os"
	//	"strconv"
	//	"strings"
	"time"
)

func (s *systemtestSuite) TestACIMode(c *C) {
	if s.fwdMode == "routing" || s.basicInfo.Scheduler == "k8" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		FwdMode:          "bridge",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
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
		FwdMode:          "bridge",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
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

		containers, err := s.runContainersInGroups(s.basicInfo.Containers, s.infoGlob.Network, s.infoGlob.Tenant, []string{"epga"})
		c.Assert(err, IsNil)

		for key := range containers {
			containersA = append(containersA, key)
		}

		// Verify containers in A can ping default gateway
		c.Assert(s.pingTestByName(containersA, s.infoGlob.Gateway), IsNil)

		c.Assert(s.removeContainers(containersA), IsNil)
		containersA = nil
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
		FwdMode:          "bridge",
		Vlans:            s.infoGlob.Vlan,
		Vxlans:           s.infoGlob.Vxlan,
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.infoGlob.Tenant,
	}), IsNil)

	containersA := []*container{}
	containersB := []*container{}

	containersA2 := []*container{}
	containersB2 := []*container{}
	for i := 0; i < s.basicInfo.Iterations; i++ {
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
			GroupName:   "epga",
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.infoGlob.Tenant,
			NetworkName: s.infoGlob.Network,
			GroupName:   "epgb",
		}), IsNil)

		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.infoGlob.Tenant,
			EndpointGroups: []string{"epga", "epgb"},
			AppProfileName: "profile1",
		}), IsNil)

		time.Sleep(5 * time.Second)

		groups := []string{"epga", "epgb"}
		containers, err := s.runContainersInGroups(s.basicInfo.Containers, s.infoGlob.Network, s.infoGlob.Tenant, groups)
		c.Assert(err, IsNil)
		time.Sleep(time.Second * 20)
		for key, value := range containers {
			if value == "epga" {
				containersA = append(containersA, key)
			} else {
				containersB = append(containersB, key)
			}
		}

		// Verify containers in epga can ping default gateway
		c.Assert(s.pingTestByName(containersA, s.infoGlob.Gateway), IsNil)
		// Verify containers in epga cannot ping containers in epgb
		c.Assert(s.pingFailureTest(containersA, containersB), IsNil)
		// Verify containers in epgb can ping default gateway
		c.Assert(s.pingTestByName(containersB, s.infoGlob.Gateway), IsNil)

		// Create a policy that allows ICMP and apply between A and B
		c.Assert(s.cli.PolicyPost(&client.Policy{
			PolicyName: "policyAB",
			TenantName: s.infoGlob.Tenant,
		}), IsNil)

		c.Assert(s.cli.RulePost(&client.Rule{
			RuleID:            "1",
			PolicyName:        "policyAB",
			TenantName:        s.infoGlob.Tenant,
			FromEndpointGroup: "epga",
			Direction:         "in",
			Protocol:          "icmp",
			Action:            "allow",
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.infoGlob.Tenant,
			NetworkName: s.infoGlob.Network,
			Policies:    []string{"policyAB"},
			GroupName:   "epgb",
		}), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epgb",
			containersB), IsNil)

		// Verify containers in epga can ping containers in epgb
		for _, cB := range containersB {}
			c.Assert(s.pingTestByName(containersA, cB.eth0.ip), IsNil)
		}

		// Verify TCP is not allowed.

		c.Assert(s.startListeners(containersA, []int{8000, 8001}), IsNil)
		c.Assert(s.startListeners(containersB, []int{8000, 8001}), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)

		c.Assert(s.cli.RulePost(&client.Rule{
			RuleID:            "2",
			PolicyName:        "policyAB",
			TenantName:        s.infoGlob.Tenant,
			FromEndpointGroup: "epga",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8000,
			Action:            "allow",
		}), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epgb",
			containersB), IsNil)

		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8001, 1, 3), IsNil)
		for _, cB := range containersB {
			c.Assert(s.pingTestByName(containersA, cB.eth0.ip), IsNil)
		}

		// Add a rule to allow 8001
		c.Assert(s.cli.RulePost(&client.Rule{
			RuleID:            "3",
			PolicyName:        "policyAB",
			TenantName:        s.infoGlob.Tenant,
			FromEndpointGroup: "epga",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8001,
			Action:            "allow",
		}), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epgb",
			containersB), IsNil)

		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8001, 1, 3), IsNil)
		for _, cB := range containersB {
			log.Infof("IP of me is %s", cB.eth0.ip)
			c.Assert(s.pingTestByName(containersA, cB.eth0.ip), IsNil)
		}

		// Delete ICMP rule
		c.Assert(s.cli.RuleDelete(s.infoGlob.Tenant, "policyAB", "1"), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epgb",
			containersB), IsNil)

		c.Assert(s.pingFailureTest(containersA, containersB), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8001, 1, 3), IsNil)
		// Delete TCP 8000 rule
		c.Assert(s.cli.RuleDelete(s.infoGlob.Tenant, "policyAB", "2"), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile1",
			"epgb",
			containersB), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA, containersB, 8001, 1, 3), IsNil)
		c.Assert(s.pingFailureTest(containersA, containersB), IsNil)

		// Delete the app profile
		c.Assert(s.cli.AppProfileDelete(s.infoGlob.Tenant, "profile1"), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA, containersB, 8001, 1, 3), IsNil)
		c.Assert(s.pingFailureTest(containersA, containersB), IsNil)

		// Create the app profile with a different name
		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.infoGlob.Tenant,
			EndpointGroups: []string{"epga", "epgb"},
			AppProfileName: "profile2",
		}), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile2",
			"epga",
			containersA), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile2",
			"epgb",
			containersB), IsNil)
		c.Assert(s.removeContainers(append(containersA, containersB...)), IsNil)
		containersA = nil
		containersB = nil
		containers, err = s.runContainersInGroups(s.basicInfo.Containers, s.infoGlob.Network, s.infoGlob.Tenant, groups)
		c.Assert(err, IsNil)
		time.Sleep(time.Second * 20)
		for key, value := range containers {
			if value == "epga" {
				containersA2 = append(containersA, key)
			} else {
				containersB2 = append(containersB, key)
			}
		}
		c.Assert(s.startListeners(containersA2, []int{8000, 8001}), IsNil)
		c.Assert(s.startListeners(containersB2, []int{8000, 8001}), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA2, containersB2, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA2, containersB2, 8001, 1, 3), IsNil)
		c.Assert(s.pingFailureTest(containersA2, containersB2), IsNil)

		// Add TCP 8000 rule.
		c.Assert(s.cli.RulePost(&client.Rule{
			RuleID:            "2",
			PolicyName:        "policyAB",
			TenantName:        s.infoGlob.Tenant,
			FromEndpointGroup: "epga",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8000,
			Action:            "allow",
		}), IsNil)
		err = errors.New("Forced")
		//c.Assert(err, IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile2",
			"epga",
			containersA2), IsNil)

		c.Assert(s.checkACILearning(s.infoGlob.Tenant,
			"profile2",
			"epgb",
			containersB2), IsNil)

		c.Assert(s.checkConnectionPairRetry(containersA2, containersB2, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(containersA2, containersB2, 8001, 1, 3), IsNil)
		c.Assert(s.pingFailureTest(containersA2, containersB2), IsNil)

		// Delete the app profile
		c.Assert(s.cli.AppProfileDelete(s.infoGlob.Tenant, "profile2"), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(s.checkNoConnectionPairRetry(containersA2, containersB2, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(containersA2, containersB2, 8001, 1, 3), IsNil)
		c.Assert(s.pingFailureTest(containersA2, containersB2), IsNil)

		c.Assert(s.removeContainers(append(containersA2, containersB2...)), IsNil)
		containersA2 = nil
		containersB2 = nil
		c.Assert(s.cli.EndpointGroupDelete(s.infoGlob.Tenant, "epga"), IsNil)
		c.Assert(s.cli.EndpointGroupDelete(s.infoGlob.Tenant, "epgb"), IsNil)
		c.Assert(s.cli.RuleDelete(s.infoGlob.Tenant, "policyAB", "2"), IsNil)
		c.Assert(s.cli.RuleDelete(s.infoGlob.Tenant, "policyAB", "3"), IsNil)

		c.Assert(s.cli.PolicyDelete(s.infoGlob.Tenant, "policyAB"), IsNil)
		c.Assert(s.cli.NetworkDelete(s.infoGlob.Tenant, s.infoGlob.Network), IsNil)
	}
}
