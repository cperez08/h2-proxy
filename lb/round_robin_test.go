package lb

import (
	"testing"

	"github.com/cperez08/h2-proxy/conn"
)

func TestPickConnectionFromRoundRobinBalancer(t *testing.T) {
	b := GetBalancer(RoundRobin)
	if _, ok := b.(*RoundRobinLB); !ok {
		t.Log("should return round robin balancer")
		t.Fail()
	}

	pool := []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    true,
			IsConnected: true,
		},
		{
			Address:     "localhost:8080",
			IsActive:    true,
			IsConnected: true,
		},
		{
			Address:     "localhost:8090",
			IsActive:    true,
			IsConnected: true,
		},
	}

	c := b.PickConnection(pool)
	if c != pool[0] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	c = b.PickConnection(pool)
	if c != pool[1] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	c = b.PickConnection(pool)
	if c != pool[2] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	// Reset balancer
	b = GetBalancer(RoundRobin)

	pool = []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    true,
			IsConnected: false,
		},
		{
			Address:     "localhost:8080",
			IsActive:    false,
			IsConnected: true,
		},
		{
			Address:     "localhost:8090",
			IsActive:    true,
			IsConnected: true,
		},
	}

	c = b.PickConnection(pool)
	if c != pool[2] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	// Reset balancer
	b = NewRoundRobin()

	pool = []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    true,
			IsConnected: false,
		},
		{
			Address:     "localhost:8080",
			IsActive:    false,
			IsConnected: true,
		},
		{
			Address:     "localhost:8090",
			IsActive:    false,
			IsConnected: false,
		},
	}

	c = b.PickConnection(pool)
	if c != nil {
		t.Log("return unexpected connection")
		t.Fail()
	}

	b.RebuildBalancer(pool)

	pool = []*conn.Connection{}

	b.RebuildBalancer(pool)
}
