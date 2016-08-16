package integration

import (
	//"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"

	. "gopkg.in/check.v1"
)

// TestTenantCreateDelete test tenant create and delete ops
func (its *integTestSuite) TestTenantCreateDelete(c *C) {
	// Create a tenant
	c.Assert(its.client.TenantPost(&client.Tenant{
		TenantName: "TestTenant",
	}), IsNil)

	err := its.client.NetworkPost(&client.Network{
		TenantName:  "TestTenant",
		NetworkName: "TestNet",
		Subnet:      "10.1.1.0/24",
		Encap:       its.encap,
	})
	assertNoErr(err, c, "creating network")

	// verify tenant state is correct
	insp, err := its.client.TenantInspect("TestTenant")
	assertNoErr(err, c, "inspecting tenant")
	log.Infof("Inspecting tenant: %+v", insp)
	c.Assert(len(insp.Oper.Endpoints), Equals, 0)
	c.Assert(insp.Oper.NumEndpoints, Equals, 0)
	c.Assert(insp.Oper.NumNet, Equals, 1)

	err = its.client.NetworkPost(&client.Network{
		TenantName:  "default",
		NetworkName: "DefaultNet",
		Subnet:      "10.1.1.0/24",
		Encap:       its.encap,
	})
	assertNoErr(err, c, "creating network")

	// verify tenant state is correct
	insp, err = its.client.TenantInspect("default")
	assertNoErr(err, c, "inspecting tenant")
	log.Infof("Inspecting tenant: %+v", insp)
	c.Assert(len(insp.Oper.Endpoints), Equals, 0)
	c.Assert(insp.Oper.NumEndpoints, Equals, 0)
	c.Assert(insp.Oper.NumNet, Equals, 1)

	for i := 0; i < its.iterations; i++ {
		addr, err := its.allocAddress("", "TestNet.TestTenant", "")
		assertNoErr(err, c, "allocating address")
		c.Assert(addr, Equals, "10.1.1.1")
		epCfg1, err := its.createEndpoint("TestTenant", "TestNet", "", addr, "")
		assertNoErr(err, c, "creating endpoint")
		insp, err := its.client.TenantInspect("TestTenant")
		assertNoErr(err, c, "inspecting tenant")
		log.Infof("Inspecting tenant: %+v", insp)
		c.Assert(insp.Oper.NumEndpoints, Equals, 1)
		
		addr, err = its.allocAddress("", "DefaultNet.default", "")
		assertNoErr(err, c, "allocating address")
		c.Assert(addr, Equals, "10.1.1.1")
		epCfg2, err := its.createEndpoint("default", "DefaultNet", "", addr, "")
		assertNoErr(err, c, "creating endpoint")
		insp, err = its.client.TenantInspect("default")
		assertNoErr(err, c, "inspecting tenant")
		log.Infof("Inspecting tenant: %+v", insp)
		c.Assert(insp.Oper.NumEndpoints, Equals, 1)

		epCfg3, err := its.createEndpoint("default", "DefaultNet", "", addr, "")
                assertNoErr(err, c, "creating endpoint")
                insp, err = its.client.TenantInspect("default")
                assertNoErr(err, c, "inspecting tenant")
                log.Infof("Inspecting tenant: %+v", insp)
                c.Assert(insp.Oper.NumEndpoints, Equals, 2)

		err = its.deleteEndpoint("TestTenant", "TestNet", "", epCfg1)
		assertNoErr(err, c, "deleting endpoint")

		err = its.deleteEndpoint("default", "DefaultNet", "", epCfg2)
		assertNoErr(err, c, "deleting endpoint")
		
		err = its.deleteEndpoint("default", "DefaultNet", "", epCfg3)
                assertNoErr(err, c, "deleting endpoint")
	}
	//assertNoErr(its.client.NetworkDelete("default", "DefaultNet"), c, "deleting network")
	assertNoErr(its.client.NetworkDelete("TestTenant", "TestNet"), c, "deleting network")

	/*for i := 0; i < its.iterations; i++ {

		addr, err := its.allocAddress("", "DefaultNet.default", "")
		assertNoErr(err, c, "allocating address")
		c.Assert(addr, Equals, "10.1.1.1")
		epCfg2, err := its.createEndpoint("default", "DefaultNet", "", addr, "")
		assertNoErr(err, c, "creating endpoint")
		insp, err := its.client.TenantInspect("default")
		assertNoErr(err, c, "inspecting tenant")
		log.Infof("Inspecting tenant: %+v", insp)
		c.Assert(insp.Oper.NumEndpoints, Equals, 1)

		err = its.deleteEndpoint("default", "DefaultNet", "", epCfg2)
		assertNoErr(err, c, "deleting endpoint")

	}*/
	assertNoErr(its.client.NetworkDelete("default", "DefaultNet"), c, "deleting network")

}
