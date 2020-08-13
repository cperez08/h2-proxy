package lb

import (
	"math/rand"

	"github.com/cperez08/h2-proxy/conn"
)

type RandomLB struct{}

// Balance return the already balanced connection
func (l *RandomLB) PickConnection(pool []*conn.Connection) *conn.Connection {
	// avoid error in rand.Intn with 0 value
	if len(pool) == 0 {
		return nil
	}

	for i := 0; i < MaxRetries; i++ {
		r := rand.Intn(len(pool))
		if pool[r].IsActive && pool[r].IsConnected {
			return pool[r]
		}
	}
	return nil
}

// RebuildBalancer useless for this balancer
func (l *RandomLB) RebuildBalancer(pool []*conn.Connection) {}
