package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"go.uber.org/zap"
	"golang.org/x/net/dns/dnsmessage"
)

var _ discovery.Scanner = (*Scanner)(nil)

const (
	discoveryQueryUdp    = "_services._dns-sd._udp.local."
	mdnsMulticastAddress = "224.0.0.251"
	mdnsPort             = 5353
	queryTimeout         = 2 * time.Second
	discoveryTimeout     = 5 * time.Second
)

type Scanner struct{}

func (s *Scanner) Name() string {
	return "mdns"
}

func (s *Scanner) Scan(ctx context.Context, out chan<- discovery.Device) error {
	log := zap.L()

	conn, mAddr, err := s.setupConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := s.sendQuery(conn, mAddr, discoveryQueryUdp); err != nil {
		return fmt.Errorf("send discovery query: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(queryTimeout)); err != nil {
		return fmt.Errorf("set initial deadline: %w", err)
	}

	discoveredServices := make(map[string]bool)
	buf := make([]byte, 8192)
	initialPhase := true
	initialDeadline := time.Now().Add(queryTimeout)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if s.isTimeout(err) {
				if initialPhase && time.Now().After(initialDeadline) {
					// Initial discovery phase done, now query discovered services
					initialPhase = false
					for service := range discoveredServices {
						s.sendQuery(conn, mAddr, service)
					}
					// Set final deadline
					if err := conn.SetReadDeadline(time.Now().Add(queryTimeout)); err != nil {
						return fmt.Errorf("set final deadline: %w", err)
					}
					continue
				}
				return nil
			}
			return fmt.Errorf("read mdns: %w", err)
		}

		device := s.processPacket(buf[:n], src, log)
		if device != nil {
			out <- *device
		}

		// Collect new services from initial discovery phase
		if initialPhase {
			for _, ans := range s.parseAnswers(buf[:n], log) {
				if ptr, ok := ans.Body.(*dnsmessage.PTRResource); ok {
					if ans.Header.Name.String() == discoveryQueryUdp {
						ptrValue := ptr.PTR.String()
						if !discoveredServices[ptrValue] {
							discoveredServices[ptrValue] = true
							log.Debug("mdns discovered service type",
								zap.String("service", ptrValue))
						}
					}
				}
			}
		}
	}
}

func (s *Scanner) setupConnection() (*net.UDPConn, *net.UDPAddr, error) {
	mAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", mdnsMulticastAddress, mdnsPort))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve mdns addr: %w", err)
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("listen udp: %w", err)
	}

	return conn, mAddr, nil
}

func (s *Scanner) sendQuery(conn *net.UDPConn, addr *net.UDPAddr, name string) error {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:               0,
			RecursionDesired: false,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  dnsmessage.MustNewName(name),
				Type:  dnsmessage.TypePTR,
				Class: dnsmessage.ClassINET,
			},
		},
	}

	buf, err := msg.Pack()
	if err != nil {
		return fmt.Errorf("pack query: %w", err)
	}

	_, err = conn.WriteToUDP(buf, addr)
	return err
}

func (s *Scanner) isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func (s *Scanner) parseAnswers(packet []byte, log *zap.Logger) []dnsmessage.Resource {
	msg, err := s.parseDNSMessage(packet)
	if err != nil {
		log.Debug("failed to parse mdns response", zap.Error(err))
		return nil
	}

	if !msg.Header.Response {
		return nil
	}

	return msg.Answers
}

func (s *Scanner) parseDNSMessage(packet []byte) (*dnsmessage.Message, error) {
	var msg dnsmessage.Message
	err := msg.Unpack(packet)
	return &msg, err
}

func (s *Scanner) processPacket(packet []byte, src *net.UDPAddr, log *zap.Logger) *discovery.Device {
	answers := s.parseAnswers(packet, log)
	if answers == nil {
		return nil
	}

	device := &discovery.Device{
		IP:       src.IP,
		Services: make(map[string]int),
		Sources:  map[string]struct{}{"mdns": {}},
		LastSeen: time.Now(),
	}

	for _, ans := range answers {
		log.Debug("mdns answer",
			zap.String("name", ans.Header.Name.String()),
			zap.String("type", ans.Header.Type.String()))

		if ptr, ok := ans.Body.(*dnsmessage.PTRResource); ok {
			ptrValue := ptr.PTR.String()
			// Log all PTR records
			log.Debug("mdns ptr answer",
				zap.String("name", ans.Header.Name.String()),
				zap.String("ptr", ptrValue))

			if ans.Header.Name.String() != discoveryQueryUdp {
				s.processServicePTR(ptrValue, device)
			}
		} else if srv, ok := ans.Body.(*dnsmessage.SRVResource); ok {
			s.processSRV(srv, device, log)
		} else if txt, ok := ans.Body.(*dnsmessage.TXTResource); ok {
			s.processTXT(txt, device)
		} else {
			log.Debug("unhandled mdns answer type",
				zap.String("name", ans.Header.Name.String()),
				zap.String("type", ans.Header.Type.String()))
		}
	}

	if len(device.Services) > 0 || device.Hostname != "" {
		return device
	}
	return nil
}

func (s *Scanner) processServicePTR(ptr string, device *discovery.Device) {
	if device.Hostname == "" {
		device.Hostname = strings.TrimSuffix(ptr, ".local.")
		device.Hostname = strings.TrimSuffix(device.Hostname, ".")
	}

	serviceName := s.extractServiceName(ptr)
	if serviceName != "" {
		device.Services[serviceName] = 0
	}
}

func (s *Scanner) processSRV(srv *dnsmessage.SRVResource, device *discovery.Device, log *zap.Logger) {
	port := int(srv.Port)
	target := srv.Target.String()

	log.Debug("mdns srv answer",
		zap.String("target", target),
		zap.Uint16("port", srv.Port))

	if device.Hostname == "" {
		device.Hostname = strings.TrimSuffix(target, ".local.")
		device.Hostname = strings.TrimSuffix(device.Hostname, ".")
	}

	serviceName := s.extractServiceNameFromTarget(target)
	if serviceName != "" {
		device.Services[serviceName] = port
	}
}

func (s *Scanner) processTXT(txt *dnsmessage.TXTResource, device *discovery.Device) {
	for _, t := range txt.TXT {
		text := string(t)
		if strings.HasPrefix(text, "model=") {
			device.Model = strings.TrimPrefix(text, "model=")
		} else if strings.HasPrefix(text, "manufacturer=") {
			device.Manufacturer = strings.TrimPrefix(text, "manufacturer=")
		}
	}
}

func (s *Scanner) extractServiceName(ptr string) string {
	parts := strings.Split(ptr, ".")
	if len(parts) < 2 {
		return ""
	}

	serviceType := parts[0]
	if strings.HasPrefix(serviceType, "_") {
		return strings.TrimPrefix(serviceType, "_")
	}
	return serviceType
}

func (s *Scanner) extractServiceNameFromTarget(target string) string {
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return ""
	}

	serviceType := parts[0]
	if strings.HasPrefix(serviceType, "_") {
		return strings.TrimPrefix(serviceType, "_")
	}
	return serviceType
}
