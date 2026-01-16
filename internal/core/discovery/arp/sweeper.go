package arp

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	maxConcurrentTriggers = 200
	triggerDeadline       = 1200 * time.Millisecond
	tcpDialTimeout        = 300 * time.Millisecond
)

var (
	udpTriggerPorts = []int{9, 33434}
	tcpTriggerPorts = []int{80, 443}
)

// Sweeper triggers ARP resolution to populate the OS ARP cache.
// Because whosthere is designed to not run with elevated privileges,
// it cannot send ARP requests directly. Instead, it triggers ARP resolution
// by sending UDP/TCP packets to IPs in the target subnet. This causes the OS
// to send ARP requests for those IPs, populating the ARP cache which can
// then be read by the ARP scanner.
type Sweeper interface {
	Start(ctx context.Context)
	Trigger(subnet *net.IPNet)
}

type sweeper struct {
	interval time.Duration
	debounce time.Duration

	logger *zap.Logger

	mu       sync.Mutex
	started  bool
	inFlight map[string]time.Time
	workCh   chan *net.IPNet
}

func NewSweeper(interval, debounce time.Duration) Sweeper {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if debounce <= 0 {
		debounce = 60 * time.Second
	}
	logger := zap.L().With(
		zap.String("scanner", "arp"),
		zap.String("component", "sweeper"),
	)
	return &sweeper{
		interval: interval,
		debounce: debounce,
		logger:   logger,
		inFlight: make(map[string]time.Time),
		workCh:   make(chan *net.IPNet, 8),
	}
}

func (s *sweeper) Start(ctx context.Context) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.mu.Unlock()

	localIP, subnet, err := getLocalNetwork()
	if err == nil && is24(subnet) {
		s.enqueue(subnet)
	}

	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if subnet != nil && is24(subnet) {
					s.enqueue(subnet)
				}
			case sn := <-s.workCh:
				if sn == nil {
					continue
				}
				s.runSweep(ctx, sn, localIP)
			}
		}
	}()
}

func (s *sweeper) Trigger(subnet *net.IPNet) {
	if subnet == nil || !is24(subnet) {
		return
	}
	s.enqueue(subnet)
}

func (s *sweeper) enqueue(subnet *net.IPNet) {
	key := subnet.String()
	now := time.Now()

	s.mu.Lock()
	last, ok := s.inFlight[key]
	if ok && now.Sub(last) < s.debounce {
		s.mu.Unlock()
		return
	}
	s.inFlight[key] = now
	s.mu.Unlock()

	select {
	case s.workCh <- subnet:
	default:
		// Drop if the queue is full to avoid blocking callers.
	}
}

func (s *sweeper) runSweep(ctx context.Context, subnet *net.IPNet, localIP net.IP) {
	ips := generateSubnetIPs(subnet, localIP)
	if len(ips) == 0 {
		return
	}

	s.logger.Info("Triggering ARP requests for subnet", zap.String("subnet", subnet.String()))
	triggerSubnetSweep(ctx, ips)
	s.logger.Debug("ARP triggering completed", zap.String("subnet", subnet.String()))
}

func is24(subnet *net.IPNet) bool {
	if subnet == nil {
		return false
	}
	ones, bits := subnet.Mask.Size()
	return ones == 24 && bits == 32
}

func triggerSubnetSweep(ctx context.Context, ips []net.IP) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentTriggers)

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(targetIP net.IP) {
			defer wg.Done()
			defer func() { <-sem }()
			sendARPTarget(targetIP)
		}(ip)
	}

	wg.Wait()
}

func sendARPTarget(ip net.IP) {
	deadline := time.Now().Add(triggerDeadline)

	for _, p := range udpTriggerPorts {
		addr := &net.UDPAddr{IP: ip, Port: p}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			continue
		}
		_ = conn.SetWriteDeadline(deadline)
		_, _ = conn.Write([]byte{0})
		_ = conn.Close()
	}

	for _, p := range tcpTriggerPorts {
		addr := net.JoinHostPort(ip.String(), strconv.Itoa(p))
		c, err := net.DialTimeout("tcp", addr, tcpDialTimeout)
		if err == nil {
			_ = c.Close()
		}
	}
}
