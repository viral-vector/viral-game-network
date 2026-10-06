//go:build integration

package support

import (
	"github.com/alicebob/miniredis/v2"
	"os"
	"testing"
	"viral-game-network/src/database"
)

func Storage(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	endpoint := os.Getenv("VGN_TEST_STORE_ENDPOINT")
	if endpoint == "" {
		t.Fatal("integration tests require VGN_TEST_STORE_ENDPOINT; see TESTING.md")
	}
	// Never read production STORE_* credentials or reuse its namespace.
	if err := database.Connect(endpoint, "vgn_test_"+database.GetUUID(), "vgn", "test", "test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)
	return Redis(t)
}
