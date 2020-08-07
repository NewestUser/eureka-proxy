package netutil

import (
	"fmt"
	"net"
	"os"
)

func OutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func Hostname() string {

	name, err := os.Hostname()

	if err != nil {
		panic(fmt.Errorf("could not get the hostname err: %s", err.Error()))
	}

	return name
}
