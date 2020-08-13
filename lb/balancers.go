package lb

import "log"

// Balancer custom balancer type
type Balancer string

// all available balancer
const (
	None       Balancer = "none"
	RoundRobin Balancer = "round_robin"
	Random     Balancer = "random"
)

// GetBalancer returns a new balancer instance
func GetBalancer(b Balancer) LoadBalancer {
	switch b {
	case None:
		return &NoBalancer{}
	case RoundRobin:
		return &RoundRobinLB{}
	case Random:
		return &RandomLB{}
	default:
		log.Println("invalid balancer ", b, " defult set up")
		return &NoBalancer{}
	}
}
