package lb

import "log"

type Balancer string

const (
	None       Balancer = "none"
	RoundRobin Balancer = "round_robin"
	Random     Balancer = "random"
)

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
