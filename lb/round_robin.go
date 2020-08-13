package lb

import (
	"math/rand"
	"sync"

	"github.com/cperez08/h2-proxy/conn"
)

// RoundRobinLB is the round robin load balancer
// which tries to balance the request evenly
// for more information visit https://en.wikipedia.org/wiki/Round-robin_DNS
type RoundRobinLB struct {
	m    sync.Mutex
	next int
}

// NewRoundRobin returns a new roundRobin instance
func NewRoundRobin() LoadBalancer {
	return &RoundRobinLB{}
}

// PickConnection picks an existing connection balanced by round robin alg
func (l *RoundRobinLB) PickConnection(pool []*conn.Connection) *conn.Connection {
	l.m.Lock()
	defer l.m.Unlock()
	for i := 0; i < MaxRetries; i++ {
		p := pool[l.next]
		l.next = (l.next + 1) % len(pool)
		if p.IsActive && p.IsConnected {
			return p
		}
	}
	return nil
}

// RebuildBalancer refresh the next connection to have a better balancing
func (l *RoundRobinLB) RebuildBalancer(pool []*conn.Connection) {
	l.m.Lock()
	defer l.m.Unlock()
	if len(pool) > 0 {
		l.next = rand.Intn(len(pool))
	}
}
