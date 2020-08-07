package fake

import (
	"fmt"
	"github.com/newestuser/eureka-proxy/lib/netutil"
)

func SingleLocalApp(appID string, port int) *Application {

	host := netutil.Hostname()
	instanceID := fmt.Sprintf("%s:%s:%d", host, appID, port)

	return SingleLocalAppWithInstance(appID, instanceID, port)
}

func SingleLocalAppWithInstance(appID, instanceID string, port int) *Application {
	host := netutil.Hostname()
	ip := netutil.OutboundIP().String()

	return SingleInstanceApp(appID, instanceID, ip, host, port)
}