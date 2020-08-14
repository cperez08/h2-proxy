package lb

import (
	"testing"

	"github.com/cperez08/h2-proxy/conn"
)

func TestPickConnectionFromNoBalancer(t *testing.T) {
	b := GetBalancer(None)
	if _, ok := b.(*NoBalancer); !ok {
		t.Log("should return none balancer")
		t.Fail()
	}

	pool := []*conn.Connection{
		{
			Address:     "localhost:8070",
			IsActive:    true,
			IsConnected: false,
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
	if c != pool[1] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	pool[0].IsActive = false

	c = b.PickConnection(pool)
	if c != pool[1] {
		t.Log("return unexpected connection")
		t.Fail()
	}

	pool[0].IsActive = true
	pool[0].IsConnected = true

	c = b.PickConnection(pool)
	if c != pool[0] {
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

	b.RebuildBalancer(pool) // increase coverage this does not do anything
}
