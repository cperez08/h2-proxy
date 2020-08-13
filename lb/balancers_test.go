package lb

import "testing"

func TestGetBalancer(t *testing.T) {
	b := GetBalancer(None)
	if _, ok := b.(*NoBalancer); !ok {
		t.Log("should return none balancer")
		t.Fail()
	}

	b = GetBalancer(Random)
	if _, ok := b.(*RandomLB); !ok {
		t.Log("should return random balancer")
		t.Fail()
	}

	b = GetBalancer(RoundRobin)
	if _, ok := b.(*RoundRobinLB); !ok {
		t.Log("should return round robin balancer")
		t.Fail()
	}

	b = GetBalancer(Balancer("other"))
	if _, ok := b.(*NoBalancer); !ok {
		t.Log("should return no balancer")
		t.Fail()
	}
}
