//go:build integration

package repository

import (
	"testing"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestFilteredListTotalsMatchResults(t *testing.T) {
	support.Storage(t)
	for _, name := range []string{"Matchable", "Other"} {
		if _, err := PutUser(&dbtype.User{Name: name}); err != nil {
			t.Fatal(err)
		}
		if _, err := PutAdmin(&dbtype.Admin{Name: name, Email: name + "@example.invalid", Password: "unused", Phone: name}); err != nil {
			t.Fatal(err)
		}
		app, err := PutApplication(&dbtype.Application{Name: name, Guid: name, Image: "game", Port: "4000", Command: "server", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := PutLobby(&dbtype.Lobby{Name: name}, app, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := PutServer(&dbtype.Server{Name: name, Guid: name, Status: "Pending", Port: 30000}, nil); err != nil {
			t.Fatal(err)
		}
	}
	checks := map[string]func(int, int, string) (int, int, error){
		"users": func(size, page int, search string) (int, int, error) {
			rows, total, err := AllUser(size, page, search)
			return len(rows), total, err
		},
		"admins": func(size, page int, search string) (int, int, error) {
			rows, total, err := AllAdmin(size, page, search)
			return len(rows), total, err
		},
		"applications": func(size, page int, search string) (int, int, error) {
			rows, total, err := AllApplication(size, page, search)
			return len(rows), total, err
		},
		"lobbies": func(size, page int, search string) (int, int, error) {
			rows, total, err := AllLobby(size, page, search)
			return len(rows), total, err
		},
		"servers": func(size, page int, search string) (int, int, error) {
			rows, total, err := AllServer(size, page, search)
			return len(rows), total, err
		},
	}
	for name, list := range checks {
		t.Run(name, func(t *testing.T) {
			for _, input := range []struct {
				search            string
				page, rows, total int
			}{{"Matchable", 1, 1, 1}, {"match", 1, 1, 1}, {"Mtbl", 1, 1, 1}, {"Matchable", 2, 0, 1}, {"NoSuchRecord", 1, 0, 0}, {"", 1, 1, 2}} {
				rows, total, err := list(1, input.page, input.search)
				if err != nil || rows != input.rows || total != input.total {
					t.Errorf("search=%q page=%d: rows=%d total=%d error=%v; want rows=%d total=%d", input.search, input.page, rows, total, err, input.rows, input.total)
				}
			}
		})
	}
}

func TestProvisioningListTotalsExcludeRunningLobbies(t *testing.T) {
	support.Storage(t)
	app, err := PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Port: "4000", Command: "server", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := PutLobby(&dbtype.Lobby{Name: "Running"}, app, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PutLobby(&dbtype.Lobby{Name: "Waiting"}, app, nil); err != nil {
		t.Fatal(err)
	}
	server, err := PutServer(&dbtype.Server{Name: "Running", Guid: "running", Status: "Running"}, ready)
	if err != nil {
		t.Fatal(err)
	}
	if err := LinkLobbyServer(ready, server); err != nil {
		t.Fatal(err)
	}
	rows, total, err := AllLobbyNotRunning(10, 1)
	if err != nil || len(rows) != 1 || rows[0].Name != "Waiting" || total != 1 {
		t.Fatalf("provisioning list count includes a running lobby: %+v %d %v", rows, total, err)
	}
}
