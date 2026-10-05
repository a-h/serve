package endpoints

import (
	"fmt"
	"net"
	"strconv"
)

// Endpoint is a network interface, and the URL that a client can use to reach
// a server through it.
type Endpoint struct {
	Interface string
	URL       string
}

// Get returns an Endpoint for each address of an active network
// interface that a server listening on addr accepts connections on.
func Get(scheme string, addr *net.TCPAddr) (endpoints []Endpoint, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagRunning == 0 {
			continue
		}
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to list addresses of interface %q: %w", iface.Name, err)
		}
		for _, ifaceAddr := range ifaceAddrs {
			ipNet, ok := ifaceAddr.(*net.IPNet)
			if !ok || !accepts(addr.IP, ipNet.IP) {
				continue
			}
			endpoints = append(endpoints, Endpoint{
				Interface: iface.Name,
				URL:       scheme + "://" + net.JoinHostPort(ipNet.IP.String(), strconv.Itoa(addr.Port)),
			})
		}
	}
	return endpoints, nil
}

func accepts(listenIP, ip net.IP) bool {
	if listenIP.IsUnspecified() {
		return !ip.IsLinkLocalUnicast()
	}
	return listenIP.Equal(ip)
}
