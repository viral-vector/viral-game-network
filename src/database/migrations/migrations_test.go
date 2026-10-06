package migrations

import (
	"strings"
	"testing"
)

func TestMigrationRegistry(t *testing.T) {
	seen := map[string]bool{}
	tables := map[string]bool{}
	for _, migration := range AllMigrations {
		if migration.Name == "" || seen[migration.Name] || strings.TrimSpace(migration.SQL) == "" {
			t.Fatalf("invalid migration %+v", migration)
		}
		seen[migration.Name] = true
		words := strings.Fields(migration.SQL)
		if len(words) > 2 && words[0] == "DEFINE" && words[1] == "TABLE" {
			tables[words[2]] = true
		}
	}
	for _, table := range []string{"Admin", "Application", "User", "Lobby", "Server", "System_Config", "System_Event", "Lobby_Application", "Lobby_Host", "Lobby_Server", "Lobby_Users"} {
		if !tables[table] {
			t.Errorf("missing schema for %s", table)
		}
	}
}
