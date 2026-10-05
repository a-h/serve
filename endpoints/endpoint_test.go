package endpoints

import (
	"net"
	"testing"
)

func TestAccepts(t *testing.T) {
	tests := []struct {
		name     string
		listenIP string
		ip       string
		expected bool
	}{
		{name: "IPv6 wildcard accepts IPv4 addresses", listenIP: "::", ip: "192.168.1.20", expected: true},
		{name: "IPv6 wildcard accepts IPv6 addresses", listenIP: "::", ip: "2001:db8::20", expected: true},
		{name: "IPv6 wildcard accepts loopback addresses", listenIP: "::", ip: "127.0.0.1", expected: true},
		{name: "IPv6 wildcard skips IPv4 link-local addresses", listenIP: "::", ip: "169.254.10.1", expected: false},
		{name: "IPv6 wildcard skips IPv6 link-local addresses", listenIP: "::", ip: "fe80::1", expected: false},
		{name: "IPv4 wildcard accepts IPv4 addresses", listenIP: "0.0.0.0", ip: "192.168.1.20", expected: true},
		{name: "IPv4 wildcard accepts IPv6 addresses because Go listens on both", listenIP: "0.0.0.0", ip: "2001:db8::20", expected: true},
		{name: "specific address accepts the same address", listenIP: "192.168.1.20", ip: "192.168.1.20", expected: true},
		{name: "specific address skips other addresses", listenIP: "192.168.1.20", ip: "192.168.1.21", expected: false},
		{name: "specific link-local address accepts the same address", listenIP: "fe80::1", ip: "fe80::1", expected: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accepts(net.ParseIP(tt.listenIP), net.ParseIP(tt.ip))
			if got != tt.expected {
				t.Errorf("got %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		addr     *net.TCPAddr
		expected string
	}{
		{name: "IPv4 loopback address is formatted as a URL", scheme: "http", addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, expected: "http://127.0.0.1:8080"},
		{name: "IPv6 loopback address is bracketed", scheme: "http", addr: &net.TCPAddr{IP: net.ParseIP("::1"), Port: 8080}, expected: "http://[::1]:8080"},
		{name: "scheme is used in the URL", scheme: "https", addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8443}, expected: "https://127.0.0.1:8443"},
		{name: "wildcard address includes loopback", scheme: "http", addr: &net.TCPAddr{IP: net.IPv6unspecified, Port: 8080}, expected: "http://127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints, err := Get(tt.scheme, tt.addr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var found bool
			for _, e := range endpoints {
				if e.URL == tt.expected && e.Interface != "" {
					found = true
				}
			}
			if !found {
				t.Errorf("expected %q in %v", tt.expected, endpoints)
			}
		})
	}
}
