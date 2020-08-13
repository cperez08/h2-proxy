package lb

import (
	"testing"

	"github.com/cperez08/h2-proxy/conn"
)

func TestPickConnectionFromRandomBalancer(t *testing.T) {
	b := GetBalancer(Random)
	if _, ok := b.(*RandomLB); !ok {
		t.Log("should return random balancer")
		t.Fail()
	}

	pool := []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    false,
			IsConnected: false,
		},
		{
			Address:     "localhost:8080",
			IsActive:    false,
			IsConnected: false,
		},
		{
			Address:     "localhost:8090",
			IsActive:    true,
			IsConnected: true,
		},
	}

	c := b.PickConnection(pool)
	if c != pool[2] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	// repeat the process twice to make sure grabs at least one mixed connection
	c = b.PickConnection(pool)
	if c != pool[2] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	c = b.PickConnection(pool)
	if c != pool[2] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	pool = []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    false,
			IsConnected: false,
		},
		{
			Address:     "localhost:8080",
			IsActive:    false,
			IsConnected: false,
		},
	}

	c = b.PickConnection(pool)
	if c != nil {
		t.Log("return unexpected connection")
		t.Fail()
	}

	pool = []*conn.Connection{}
	c = b.PickConnection(pool)
	if c != nil {
		t.Log("return unexpected connection")
		t.Fail()
	}
}
