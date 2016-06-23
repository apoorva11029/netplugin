package systemtests

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	. "gopkg.in/check.v1"
	"time"
)

func (s *systemtestSuite) TestACIMode(c *C) {
	if s.fwdMode == "routing" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.acinfo.Vlan,
		Vxlans:           s.acinfo.Vxlan,
	}), IsNil)

	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.acinfo.Tenant,
	}), IsNil)
	c.Assert(s.cli.NetworkPost(&client.Network{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		Subnet:      s.acinfo.Subnet,
		Gateway:     s.acinfo.Gateway,
		Encap:       s.acinfo.Encap,
	}), IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		GroupName:   "epgA",
	}), IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		GroupName:   "epgB",
	}), IsNil)

	netstrA := "epgA/" + s.acinfo.Tenant
	netstrB := "epgB/" + s.acinfo.Tenant
	cA1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
	c.Assert(err, IsNil)

	cA2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
	c.Assert(err, IsNil)

	cB1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
	c.Assert(err, IsNil)

	cB2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
	c.Assert(err, IsNil)

	// Verify cA1 can ping cA2
	c.Assert(cA1.checkPing(cA2.eth0.ip), IsNil)
	// Verify cB1 can ping cB2
	c.Assert(cB1.checkPing(cB2.eth0.ip), IsNil)
	// Verify cA1 cannot ping cB1
	c.Assert(cA1.checkPingFailure(cB1.eth0.ip), IsNil)

	c.Assert(s.removeContainers([]*container{cA1, cA2, cB1, cB2}), IsNil)
	c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgA"), IsNil)
	c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgB"), IsNil)
	c.Assert(s.cli.NetworkDelete(s.acinfo.Tenant, s.acinfo.Network), IsNil)
}

func (s *systemtestSuite) TestACIPingGateway(c *C) {
	if s.fwdMode == "routing" {
		return
	}

	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.acinfo.Vlan,
		Vxlans:           s.acinfo.Vxlan,
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.acinfo.Tenant,
	}), IsNil)
	c.Assert(s.cli.NetworkPost(&client.Network{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		Subnet:      s.acinfo.Subnet,
		Gateway:     s.acinfo.Gateway,
		Encap:       s.acinfo.Encap,
	}), IsNil)
	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		GroupName:   "epgA",
	}), IsNil)

	c.Assert(s.cli.AppProfilePost(&client.AppProfile{
		TenantName:     s.acinfo.Tenant,
		EndpointGroups: []string{"epgA"},
		AppProfileName: "profile1",
	}), IsNil)
	netstr := "epgA/" + s.acinfo.Tenant
	cA1, err := s.nodes[0].runContainer(containerSpec{networkName: netstr})
	c.Assert(err, IsNil)

	// Verify cA1 can ping default gateway
	c.Assert(cA1.checkPingWithCount("20.1.1.254", 5), IsNil)

	c.Assert(s.removeContainers([]*container{cA1}), IsNil)
	c.Assert(s.cli.AppProfileDelete(s.acinfo.Tenant, "profile1"), IsNil)
	c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgA"), IsNil)
	c.Assert(s.cli.NetworkDelete(s.acinfo.Tenant, s.acinfo.Network), IsNil)
}

func (s *systemtestSuite) TestACIProfile(c *C) {
	if s.fwdMode == "routing" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.acinfo.Vlan,
		Vxlans:           s.acinfo.Vxlan,
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.acinfo.Tenant,
	}), IsNil)

	for i := 0; i < 2; i++ {
		log.Infof(">>ITERATION #%d<<", i)
		c.Assert(s.cli.NetworkPost(&client.Network{
			TenantName:  s.acinfo.Tenant,
			NetworkName: s.acinfo.Network,
			Subnet:      s.acinfo.Subnet,
			Gateway:     s.acinfo.Gateway,
			Encap:       s.acinfo.Encap,
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.acinfo.Tenant,
			NetworkName: s.acinfo.Network,
			GroupName:   "epgA",
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.acinfo.Tenant,
			NetworkName: s.acinfo.Network,
			GroupName:   "epgB",
		}), IsNil)

		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.acinfo.Tenant,
			EndpointGroups: []string{"epgA", "epgB"},
			AppProfileName: "profile1",
		}), IsNil)
		netstrA := "epgA/" + s.acinfo.Tenant
		netstrB := "epgB/" + s.acinfo.Tenant
		time.Sleep(5 * time.Second)
		cA1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
		c.Assert(err, IsNil)

		// Verify cA1 can ping default gateway
		c.Assert(cA1.checkPingWithCount("20.1.1.254", 5), IsNil)

		cB1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
		c.Assert(err, IsNil)

		// Verify cA1 cannot ping cB1
		c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)
		// Verify cB1 can ping default gateway
		c.Assert(cB1.checkPingWithCount("20.1.1.254", 5), IsNil)

		// Create a policy that allows ICMP and apply between A and B
		c.Assert(s.cli.PolicyPost(&client.Policy{
			PolicyName: "policyAB",
			TenantName: s.acinfo.Tenant,
		}), IsNil)

		c.Assert(s.cli.RulePost(&client.Rule{
			RuleID:            "1",
			PolicyName:        "policyAB",
			TenantName:        s.acinfo.Tenant,
			FromEndpointGroup: "epgA",
			Direction:         "in",
			Protocol:          "icmp",
			Action:            "allow",
		}), IsNil)

		c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
			TenantName:  s.acinfo.Tenant,
			NetworkName: s.acinfo.Network,
			Policies:    []string{"policyAB"},
			GroupName:   "epgB",
		}), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
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
			TenantName:        s.acinfo.Tenant,
			FromEndpointGroup: "epgA",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8000,
			Action:            "allow",
		}), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
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
			TenantName:        s.acinfo.Tenant,
			FromEndpointGroup: "epgA",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8001,
			Action:            "allow",
		}), IsNil)
		//cA1.checkPing("20.1.1.254")
		//cB1.checkPing("20.1.1.254")
		time.Sleep(time.Second * 5)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgB",
			cB1), IsNil)

		c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
		c.Assert(cA1.checkPingWithCount(cB1.eth0.ip, 5), IsNil)

		// Delete ICMP rule
		c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "1"), IsNil)
		time.Sleep(time.Second * 5)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgB",
			cB1), IsNil)

		c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)
		c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)

		// Delete TCP 8000 rule
		c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "2"), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile1",
			"epgB",
			cB1), IsNil)

		c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
		c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)

		// Delete the app profile
		c.Assert(s.cli.AppProfileDelete(s.acinfo.Tenant, "profile1"), IsNil)
		time.Sleep(time.Second * 5)
		//cA1.checkPingWithCount("20.1.1.254", 5)
		//cB1.checkPingWithCount("20.1.1.254", 5)
		c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
		c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)

		// Create the app profile with a different name
		c.Assert(s.cli.AppProfilePost(&client.AppProfile{
			TenantName:     s.acinfo.Tenant,
			EndpointGroups: []string{"epgA", "epgB"},
			AppProfileName: "profile2",
		}), IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile2",
			"epgA",
			cA1), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile2",
			"epgB",
			cB1), IsNil)

		//cA1.checkPingWithCount("20.1.1.254", 5)
		//cB1.checkPingWithCount("20.1.1.254", 5)
		cA2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
		c.Assert(err, IsNil)
		cB2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
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
			TenantName:        s.acinfo.Tenant,
			FromEndpointGroup: "epgA",
			Direction:         "in",
			Protocol:          "tcp",
			Port:              8000,
			Action:            "allow",
		}), IsNil)
		err = errors.New("Forced")
		//c.Assert(err, IsNil)
		time.Sleep(time.Second * 5)
		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile2",
			"epgA",
			cA2), IsNil)

		c.Assert(checkACILearning(s.acinfo.Tenant,
			"profile2",
			"epgB",
			cB2), IsNil)

		//cA1.checkPingWithCount("20.1.1.254", 5)
		//cB1.checkPingWithCount("20.1.1.254", 5)

		c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
		c.Assert(cA2.checkPingFailureWithCount(cB2.eth0.ip, 5), IsNil)

		// Delete the app profile
		c.Assert(s.cli.AppProfileDelete(s.acinfo.Tenant, "profile2"), IsNil)
		time.Sleep(time.Second * 5)
		//cA1.checkPingWithCount("20.1.1.254", 5)
		//cB1.checkPingWithCount("20.1.1.254", 5)
		c.Assert(s.checkNoConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
		c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)
		c.Assert(cA2.checkPingFailureWithCount(cB2.eth0.ip, 5), IsNil)

		c.Assert(s.removeContainers([]*container{cA1, cB1, cA2, cB2}), IsNil)
		c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgA"), IsNil)
		c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgB"), IsNil)
		c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "2"), IsNil)
		c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "3"), IsNil)
		c.Assert(s.cli.PolicyDelete(s.acinfo.Tenant, "policyAB"), IsNil)
		c.Assert(s.cli.NetworkDelete(s.acinfo.Tenant, s.acinfo.Network), IsNil)
	}
}

func (s *systemtestSuite) TestACIGW_Restart(c *C) {
	if s.fwdMode == "routing" {
		return
	}
	c.Assert(s.cli.GlobalPost(&client.Global{
		Name:             "global",
		NetworkInfraType: "aci",
		Vlans:            s.acinfo.Vlan,
		Vxlans:           s.acinfo.Vxlan,
	}), IsNil)
	c.Assert(s.cli.TenantPost(&client.Tenant{
		TenantName: s.acinfo.Tenant,
	}), IsNil)

	c.Assert(s.cli.NetworkPost(&client.Network{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		Subnet:      s.acinfo.Subnet,
		Gateway:     s.acinfo.Gateway,
		Encap:       s.acinfo.Encap,
	}), IsNil)
	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		GroupName:   "epgA",
	}), IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		GroupName:   "epgB",
	}), IsNil)

	c.Assert(s.cli.AppProfilePost(&client.AppProfile{
		TenantName:     s.acinfo.Tenant,
		EndpointGroups: []string{"epgA", "epgB"},
		AppProfileName: "profile1",
	}), IsNil)
	netstrA := "epgA/" + s.acinfo.Tenant
	netstrB := "epgB/" + s.acinfo.Tenant
	time.Sleep(5 * time.Second)
	cA1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
	c.Assert(err, IsNil)
	// Verify cA1 can ping default gateway
	c.Assert(cA1.checkPingWithCount("20.1.1.254", 5), IsNil)

	cB1, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
	c.Assert(err, IsNil)

	// Verify cA1 cannot ping cB1
	c.Assert(cA1.checkPingFailureWithCount(cB1.eth0.ip, 5), IsNil)
	// Verify cB1 can ping default gateway
	c.Assert(cB1.checkPingWithCount("20.1.1.254", 5), IsNil)

	cA2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrA})
	c.Assert(err, IsNil)
	cB2, err := s.nodes[0].runContainer(containerSpec{networkName: netstrB})
	c.Assert(err, IsNil)

	// Verify cA1 can ping cA2
	c.Assert(cA1.checkPingWithCount(cA2.eth0.ip, 5), IsNil)
	// Verify cB1 can ping cB2
	c.Assert(cB1.checkPingWithCount(cB2.eth0.ip, 5), IsNil)

	//s.nodes[0].restartContainer("contiv-aci-gw")
	//time.Sleep(100 * time.Millisecond)
	c.Assert(s.cli.PolicyPost(&client.Policy{
		PolicyName: "policyAB",
		TenantName: s.acinfo.Tenant,
	}), IsNil)

	c.Assert(s.cli.EndpointGroupPost(&client.EndpointGroup{
		TenantName:  s.acinfo.Tenant,
		NetworkName: s.acinfo.Network,
		Policies:    []string{"policyAB"},
		GroupName:   "epgB",
	}), IsNil)
	time.Sleep(time.Second * 5)

	c.Assert(s.cli.RulePost(&client.Rule{
		RuleID:            "1",
		PolicyName:        "policyAB",
		TenantName:        s.acinfo.Tenant,
		FromEndpointGroup: "epgA",
		Direction:         "in",
		Protocol:          "icmp",
		Action:            "allow",
	}), IsNil)
	time.Sleep(time.Second * 5)
	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgA",
		cA1), IsNil)

	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgB",
		cB1), IsNil)

	c.Assert(s.cli.RulePost(&client.Rule{
		RuleID:            "2",
		PolicyName:        "policyAB",
		TenantName:        s.acinfo.Tenant,
		FromEndpointGroup: "epgA",
		Direction:         "in",
		Protocol:          "tcp",
		Port:              8000,
		Action:            "allow",
	}), IsNil)
	time.Sleep(time.Second * 5)
	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgA",
		cA1), IsNil)

	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgB",
		cB1), IsNil)

	containers := []*container{cA1, cB1}
	from := []*container{cA1}
	to := []*container{cB1}
	c.Assert(s.startListeners(containers, []int{8000, 8001}), IsNil)
	c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
	c.Assert(s.checkNoConnectionPairRetry(from, to, 8001, 1, 3), IsNil)

	s.nodes[0].restartContainers("contiv-aci-gw")
	time.Sleep(100 * time.Millisecond)
	c.Assert(s.cli.RulePost(&client.Rule{
		RuleID:            "3",
		PolicyName:        "policyAB",
		TenantName:        s.acinfo.Tenant,
		FromEndpointGroup: "epgA",
		Direction:         "in",
		Protocol:          "tcp",
		Port:              8001,
		Action:            "allow",
	}), IsNil)
	//cA1.checkPing("20.1.1.254")
	//cB1.checkPing("20.1.1.254")
	time.Sleep(time.Second * 5)

	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgA",
		cA1), IsNil)
	c.Assert(checkACILearning(s.acinfo.Tenant,
		"profile1",
		"epgB",
		cB1), IsNil)

	c.Assert(s.checkConnectionPairRetry(from, to, 8000, 1, 3), IsNil)
	c.Assert(s.checkConnectionPairRetry(from, to, 8001, 1, 3), IsNil)

	c.Assert(s.checkNoConnectionPairRetry(from, to, 8002, 1, 3), IsNil)

	c.Assert(cA1.checkPingWithCount("20.1.1.254", 5), IsNil)
	c.Assert(cA1.checkPingWithCount(cB1.eth0.ip, 5), IsNil)
	// Verify cB1 can ping default gateway
	c.Assert(cB1.checkPingWithCount("20.1.1.254", 5), IsNil)
	// Verify cA1 can ping cA2
	c.Assert(cA1.checkPingWithCount(cA2.eth0.ip, 5), IsNil)
	// Verify cB1 can ping cB2
	c.Assert(cB1.checkPingWithCount(cB2.eth0.ip, 5), IsNil)

	c.Assert(s.removeContainers([]*container{cA1, cB1, cA2, cB2}), IsNil)
	c.Assert(s.cli.AppProfileDelete(s.acinfo.Tenant, "profile1"), IsNil)
	c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgA"), IsNil)
	c.Assert(s.cli.EndpointGroupDelete(s.acinfo.Tenant, "epgB"), IsNil)
	c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "2"), IsNil)
	c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "3"), IsNil)
	c.Assert(s.cli.RuleDelete(s.acinfo.Tenant, "policyAB", "1"), IsNil)
	c.Assert(s.cli.PolicyDelete(s.acinfo.Tenant, "policyAB"), IsNil)
	c.Assert(s.cli.NetworkDelete(s.acinfo.Tenant, s.acinfo.Network), IsNil)
}
