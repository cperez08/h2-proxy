package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	r := NewResolver(5, true)

	ips := r.Resolve("localhost", "8080")
	assert.NotEqual(t, 0, len(ips))
	assert.NotEqual(t, 0, len(r.GetCurrentIPs()))
	r.CloseResolver()
}
