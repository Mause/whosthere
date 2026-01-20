package arp

import (
	"errors"
	"net"
)

var (
	ErrNoIPv4Interface = errors.New("arp: no IPv4 network interface found")
)

func generateSubnetIPs(subnet *net.IPNet, skipIP net.IP) []net.IP {
	var ips []net.IP
	network := subnet.IP.To4()
	if network == nil {
		return ips // Not IPv4
	}

	// Calculate the network and broadcast IPs
	networkIP := make(net.IP, len(network))
	copy(networkIP, network)
	broadcastIP := make(net.IP, len(network))
	copy(broadcastIP, network)
	for i := range network {
		broadcastIP[i] |= ^subnet.Mask[i]
	}

	// Start from the first host IP (network + 1)
	currentIP := incrementIP(networkIP)

	// Iterate until broadcast IP
	// TODO (ramon) add a max to this, e.g. up to /16 otherwise log and just don't go till the broadcast IP
	for !currentIP.Equal(broadcastIP) {
		if !currentIP.Equal(skipIP) {
			ipCopy := make(net.IP, len(currentIP))
			copy(ipCopy, currentIP)
			ips = append(ips, ipCopy)
		}
		currentIP = incrementIP(currentIP)
	}

	return ips
}

// incrementIP increments the IP address by 1
func incrementIP(ip net.IP) net.IP {
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)
	for i := len(newIP) - 1; i >= 0; i-- {
		newIP[i]++
		if newIP[i] != 0 {
			break
		}
	}
	return newIP
}
