//go:build integration

package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestConcurrentJoinsRespectCapacity(t *testing.T) {
	support.Storage(t)
	app, err := PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Port: "4000", Command: "server", Lobby_Max_Players: "2", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	host, err := PutUser(&dbtype.User{Name: "Host"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := PutLobby(&dbtype.Lobby{Name: "Room"}, app, host)
	if err != nil {
		t.Fatal(err)
	}
	players := make([]*dbtype.User, 12)
	for i := range players {
		players[i], err = PutUser(&dbtype.User{Name: fmt.Sprintf("Player %d", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var joined atomic.Int32
	for _, player := range players {
		wg.Add(1)
		go func(user *dbtype.User) {
			defer wg.Done()
			<-start
			if _, err := JoinLobby(lobby.ModelID(), user, ""); err == nil {
				joined.Add(1)
			}
		}(player)
	}
	close(start)
	wg.Wait()
	current, err := GetLobby(lobby.ModelID())
	if err != nil || len(current.Lobby_Users) != 2 || joined.Load() != 1 {
		t.Fatalf("capacity violated: joined=%d room=%+v error=%v", joined.Load(), current, err)
	}
	for _, member := range current.Lobby_Users {
		if member.ModelID() == players[0].ModelID() {
			players[0] = players[1]
		}
	}
	// A rejected join must leave the user's existing membership intact.
	other, err := PutLobby(&dbtype.Lobby{Name: "Other"}, app, players[0])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := JoinLobby(lobby.ModelID(), players[0], ""); err == nil {
		t.Fatal("full lobby accepted another user")
	}
	other, err = GetLobby(other.ModelID())
	if err != nil || other.Lobby_Host == nil || other.Lobby_Host.ModelID() != players[0].ModelID() {
		t.Fatalf("rejected join removed old membership: %+v %v", other, err)
	}
}

func TestLobbyApplicationPatchReplacesRelationship(t *testing.T) {
	support.Storage(t)
	apps := make([]*dbtype.Application, 2)
	for i := range apps {
		var err error
		apps[i], err = PutApplication(&dbtype.Application{Name: fmt.Sprintf("Game %d", i), Guid: fmt.Sprintf("game-%d", i), Image: "game", Port: "4000", Command: "server", Lobby_Max_Players: "2", Lobby_Max_Persist: "30"})
		if err != nil {
			t.Fatal(err)
		}
	}
	lobby, err := PutLobby(&dbtype.Lobby{Name: "Room"}, apps[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := PatchLobby(lobby.ModelID(), &LobbyPatch{Lobby_Application: apps[1]})
	if err != nil || updated.Lobby_Application == nil || updated.Lobby_Application.ModelID() != apps[1].ModelID() {
		t.Fatalf("application change not persisted: %+v %v", updated, err)
	}
}
