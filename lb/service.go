package lb

import (
	"github.com/cperez08/h2-proxy/conn"
)

// MaxRetries maximum number of attempts to get an active connections from the pool
const MaxRetries = 10

type LoadBalancer interface {
	// PickConnection returns the balanced connection according
	// to the chosen algorithm
	PickConnection(pool []*conn.Connection) *conn.Connection
	// RebuildBalancer notify the balancer about changes in the pool
	RebuildBalancer(pool []*conn.Connection)
}
