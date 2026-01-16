package arp

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrNoIPv4Interface = errors.New("arp: no IPv4 network interface found")
)

// getLocalNetwork returns local IPv4 address and subnet.
// Returns the first non-loopback IPv4 interface found.
func getLocalNetwork() (net.IP, *net.IPNet, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("list interfaces: %w", err)
	}

	for _, iface := range ifaces {
		// Skip down and loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip != nil && !ip.IsLoopback() {
				return ip, ipNet, nil
			}
		}
	}

	return nil, nil, ErrNoIPv4Interface
}

func generateSubnetIPs(subnet *net.IPNet, skipIP net.IP) []net.IP {
	var ips []net.IP
	network := subnet.IP.To4()
	if network == nil {
		return ips // Not IPv4
	}

	ones, bits := subnet.Mask.Size()
	if ones != 24 || bits != 32 {
		return ips // Not a /24 network
	}

	for i := 1; i <= 254; i++ {
		ip := net.IPv4(network[0], network[1], network[2], byte(i))
		if !ip.Equal(skipIP) {
			ips = append(ips, ip)
		}
	}

	return ips
}
