package lb

import "github.com/cperez08/h2-proxy/conn"

type NoBalancer struct{}

// Balance return the first active connection found
func (l *NoBalancer) PickConnection(pool []*conn.Connection) *conn.Connection {
	for i := 0; i < len(pool); i++ {
		if pool[i].IsActive && pool[i].IsConnected {
			return pool[i]
		}
	}

	return nil
}

// RebuildBalancer useless for this balancer
func (l *NoBalancer) RebuildBalancer(pool []*conn.Connection) {}
