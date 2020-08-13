package resolver

// Service ...
type Service interface {
	// resolve the domain returning the list of A and AAAA records
	// associated to the domain
	Resolve(domain string) []string

	// GetCurrentIPs returns the set of ips stored by the resolver
	GetCurrentIPs() []string

	// CloseResolver closes the resolver stopping watch for
	// changes in the domain
	CloseResolver()
}
