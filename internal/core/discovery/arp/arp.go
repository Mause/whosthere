package arp

import (
	"context"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"go.uber.org/zap"
)

var _ discovery.Scanner = (*Scanner)(nil)

// Scanner implements ARP-based discovery by reading the ARP cache.
// Optional Sweeper can populate the cache in the background.
type Scanner struct {
	iface   *discovery.InterfaceInfo
	Sweeper *Sweeper

	logger *zap.Logger
}

func NewScanner(iface *discovery.InterfaceInfo, sweeper *Sweeper) *Scanner {
	return &Scanner{iface: iface, Sweeper: sweeper}
}

func (s *Scanner) Name() string { return "arp" }

// Scan performs ARP discovery.
func (s *Scanner) Scan(ctx context.Context, out chan<- discovery.Device) error {
	if s.logger == nil {
		s.logger = zap.L().With(zap.String("scanner", s.Name()))
	}

	if s.Sweeper != nil {
		// this feels awkward
		s.Sweeper.Start(ctx)
	}

	time.Sleep(1 * time.Second)

	return s.readARPCache(ctx, out)
}
