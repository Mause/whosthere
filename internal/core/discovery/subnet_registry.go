package discovery

import (
	"net"
	"sync"
)

// SubnetRegistry tracks discovered subnets to avoid redundant work.
type SubnetRegistry struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func NewSubnetRegistry() *SubnetRegistry {
	return &SubnetRegistry{seen: make(map[string]struct{})}
}

// Add records the subnet and reports true if it was not seen before.
func (r *SubnetRegistry) Add(subnet *net.IPNet) bool {
	if subnet == nil {
		return false
	}
	key := subnet.String()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.seen[key]; ok {
		return false
	}
	r.seen[key] = struct{}{}
	return true
}
