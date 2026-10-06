package support

import (
	"github.com/alicebob/miniredis/v2"
	"testing"
	"viral-game-network/src/cache"
	"viral-game-network/src/pubsub"
)

// Redis gives each test an isolated server without requiring an installed Redis.
// Tests using package-level clients must not run in parallel within a package.
func Redis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server := miniredis.RunT(t)
	cache.Configure(server.Addr(), "")
	pubsub.Configure(server.Addr(), "")
	return server
}
