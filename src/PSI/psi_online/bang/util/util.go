package util

import (
	"fmt"
	"net"
	"strconv"
)

func ParserAddr(addr string) (host string, port int) {
	host, ports, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}

	port, err = strconv.Atoi(ports)
	if err != nil {
		panic(err)
	}

	return
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func ParseNodeAddr(addr string) (host string, port int) {
	host, port = ParserAddr(addr)
	host = GetLocalIP()

	return
}

func GetNodeAddr(addr string) string {
	host, port := ParseNodeAddr(addr)

	return fmt.Sprintf("%s:%d", host, port)
}

func GetIdelPort() (port int) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}

	// port, err = net.LookupPort("tcp", portStr)
	port, err = strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	return
}
