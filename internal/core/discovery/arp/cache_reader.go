package arp

import (
	"context"
	"net"
	"runtime"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

func (s *Scanner) readARPCache(ctx context.Context, out chan<- discovery.Device) error {
	switch runtime.GOOS {
	case "linux":
		return s.readLinuxARPCache(ctx, out)
	case "darwin", "freebsd", "netbsd", "openbsd":
		return s.readDarwinARPCache(ctx, out)
	default:
		return nil
	}
}

// Entry represents a single ARP cache entry.
type Entry struct {
	IP  net.IP
	MAC net.HardwareAddr
	Age time.Duration // How old the entry is (0 if unknown)
}

// emitARPEntries sends discovered ARP entries to the output channel.
func (s *Scanner) emitARPEntries(ctx context.Context, out chan<- discovery.Device, entries []Entry) error {
	now := time.Now()

	// Get local subnet once to check broadcast addresses
	_, subnet, _ := getLocalNetwork()

	for _, entry := range entries {
		if entry.IP == nil || entry.MAC == nil {
			continue
		}

		// Filter non-device addresses:
		// - skip multicast MACs (I/G bit set)
		// - skip broadcast MAC (FF:FF:FF:FF:FF:FF)
		// - skip IPv4 broadcast address for our subnet
		// - skip IPv4 multicast ranges (224.0.0.0/4)
		ip4 := entry.IP.To4()
		if isMulticastMAC(entry.MAC) || isBroadcastMAC(entry.MAC) || (ip4 != nil && ip4[0]&0xF0 == 224) || isBroadcastIPv4(entry.IP, subnet) {
			continue
		}

		dd := discovery.NewDevice(entry.IP)
		dd.MAC = entry.MAC.String()
		dd.Sources["arp"] = struct{}{}

		if entry.Age > 0 {
			dd.LastSeen = now.Add(-entry.Age)
		} else {
			dd.LastSeen = now
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- dd:
		}
	}

	return nil
}

// isMulticastMAC checks if a MAC address is a multicast address.
func isMulticastMAC(mac net.HardwareAddr) bool {
	// Multicast MACs have the least significant bit of the first byte set
	return len(mac) > 0 && (mac[0]&0x01) != 0
}

// isBroadcastMAC checks if a MAC address is a broadcast address.
func isBroadcastMAC(mac net.HardwareAddr) bool {
	// Broadcast MAC is FF:FF:FF:FF:FF:FF
	return len(mac) == 6 && mac[0] == 0xFF && mac[1] == 0xFF && mac[2] == 0xFF && mac[3] == 0xFF && mac[4] == 0xFF && mac[5] == 0xFF
}

// isBroadcastIPv4 checks if an IPv4 address is a broadcast address for the given subnet.
func isBroadcastIPv4(ip net.IP, subnet *net.IPNet) bool {
	if subnet == nil || ip == nil {
		return false
	}
	bcast := getBroadcastAddress(subnet)
	if bcast == nil {
		return false
	}
	return ip.To4() != nil && ip.Equal(bcast)
}

// getBroadcastAddress calculates the IPv4 broadcast address for a given subnet.
func getBroadcastAddress(subnet *net.IPNet) net.IP {
	if subnet == nil {
		return nil
	}
	net4 := subnet.IP.To4()
	mask := subnet.Mask
	if net4 == nil || mask == nil || len(mask) != net.IPv4len {
		return nil
	}
	return net.IPv4(
		net4[0]|^mask[0],
		net4[1]|^mask[1],
		net4[2]|^mask[2],
		net4[3]|^mask[3],
	)
}
