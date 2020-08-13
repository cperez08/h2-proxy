package resolver

import (
	"time"

	rsv "github.com/cperez08/dm-resolver/pkg/resolver"
)

// Resolver ...
type Resolver struct {
	r           *rsv.DomainResolver
	C           chan bool
	refreshRate time.Duration
	needRefresh bool
}

// NewResolver creates a new resolver instance
func NewResolver(refreshRate int, needRefresh bool) *Resolver {
	return &Resolver{
		C:           make(chan bool),
		refreshRate: time.Duration(refreshRate),
		needRefresh: needRefresh,
	}
}

// Resolve resolve the host and returns the domains associated to it
// return the address in host:port format
func (r *Resolver) Resolve(host, port string) []string {
	r.r = rsv.NewResolver(host, port, r.needRefresh, &r.refreshRate, r.C)
	r.r.StartResolver()
	return r.r.Addresses
}

// GetCurrentIPs return the current ips set in the resolver
func (r *Resolver) GetCurrentIPs() []string {
	return r.r.Addresses
}

// CloseResolver closes the resolver
func (r *Resolver) CloseResolver() {
	r.r.Close()
}
