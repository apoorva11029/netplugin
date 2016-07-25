A guide to running netplugin SYSTEMTESTS on Vagrant and Baremetal platforms:

Customize the JSON file netplugin/systemtests/cfg.json according to your environment. A typical file for vagrant with swarm looks like:
```
[
    {
      "scheduler" : "swarm",      //Scheduler used : Docker, Swarm, k8s
      "swarm_variable":"DOCKER_HOST=192.168.2.10:2375",    //Env variable for swarm. Typically <master node's IP>:2375
      "platform" : "vagrant",    //Platform: Vagrant or Platform
      "product" : "netplugin",    // Product: netplugin or volplugin(not yet supported)
      "aci_mode" : "off",      // ACI mode: on/off
      "short"   : false,      // Do a quick validation run instead of the full test suite
      "containers" : 3,       // Number of containers to use
      "iterations" : 2,       // Number of iterations
      "enableDNS" : false,     //Enable DNS service discovery
      "contiv_cluster_store" : "etcd://localhost:2379",      //cluster store URL for etcd or consul
      "contiv_l3" : "",       //For running in routing mode
      "key_file" : "",     //Insecure private key for swarm setup on Baremetal
      "binpath" : "/opt/gopath/bin",    //netplugin/netmaster binary path. /home/admin/bin for baremetal

      "hostips" : "",         // host IPs for swarm setup on Baremetal, separated by comma
      "hostusernames" : "",     // host usernames for swarm setup on Baremetal, separated by comma
      "dataInterface" : "eth2",   
      "mgmtInterface" : "eth1",

      // variables for ACI tests:
      "vlan" : "1120-1150",    
      "vxlan" : "1-10000",
      "subnet" : "10.1.1.0/24",
      "gateway" : "10.1.1.254",
      "network" : "default",
      "tenant": "TestTenant",
      "encap" : "vlan",
      "master" : true
      }
]
```

Testing with Vagrant:

1. Make a suitable JSON file on your local machine (inside the systemtests directory).
2. From the netplugin directory of your machine (outside the vagrant nodes), run:

```
  make system-test
```
Testing with Baremetal:

Scheduler: Swarm

1. Make a suitable JSON file on your local machine (inside the systemtests directory).
2. Make a suitable YML file in the same location for bringing up swarm, according to your host IPs and ACI mode. Check our sample cfg.yml for reference.  
3. Set these Environment variables on the master node:

```
export GOPATH=/home/admin
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN:/usr/local/go/bin
```

4. Build the code on master node. You can run from $GOPATH/src/github.com/contiv/netplugin
```
make run-build
```
5.  Run Systemtests like this
```
godep go test -v -timeout 240m ./systemtests -check.v -check.f "<Name of ACI Test Function>"
for eg :

godep go test -v -timeout 240m ./systemtests -check.v -check.f "TestACIMode"

	This will run TestACIMode test function metioned in systemtests/aci_test.go

godep go test -v -timeout 240m ./systemtests -check.v -check.f "TestACI"

	This will run all the test function which are Starting from TestACI
```
