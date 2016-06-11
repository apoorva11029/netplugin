package systemtests

type systemTestScheduler interface {
	runContainer(spec containerSpec) (*container, error)
	stop(c *container) error
	start(c *container) error
	startNetmaster() error
	stopNetmaster() error
	stopNetplugin() error
	startNetplugin(args string) error
	cleanupContainers() error
	checkNoConnection(c *container, ipaddr, protocol string, port int) error
	checkConnection(c *container, ipaddr, protocol string, port int) error
	startListener(c *container, port int, protocol string) error
	rm(c *container) error
	getIPAddr(c *container, dev string) (string, error)
	checkPing(c *container, ipaddr string) error
	checkPingFailure(c *container, ipaddr string) error
	cleanupSlave()
	cleanupMaster()
}
