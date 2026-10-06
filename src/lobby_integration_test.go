//go:build integration

package main

import (
	"testing"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestLobbyPrivacyAndMembership(t *testing.T) {
	redis := support.Storage(t)
	t.Setenv("CACHE_ENDPOINT", redis.Addr())
	app := NewApp("../views", "../public")
	t.Cleanup(func() { app.Shutdown() })
	player := func(name string) (*dbtype.User, string) {
		user, err := repository.PutUser(&dbtype.User{Name: name})
		if err != nil {
			t.Fatal(err)
		}
		token, err := auth.GenerateToken(name, user.ModelID(), auth.UserAudience)
		if err != nil {
			t.Fatal(err)
		}
		return user, token
	}
	host, hostToken := player("Host")
	_, guestToken := player("Guest")
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "server", Lobby_Max_Players: "2", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	room := func(name string, private bool) *dbtype.Lobby {
		lobby, err := repository.PutLobby(&dbtype.Lobby{Name: name, Private: private, Code: "room-code"}, application, host)
		if err != nil {
			t.Fatal(err)
		}
		return lobby
	}
	t.Run("join accepts code from JSON", func(t *testing.T) {
		lobby := room("Private", true)
		request(t, app, "POST", "/api/lobby/"+lobby.ModelID()+"/join", map[string]string{"code": "wrong"}, guestToken, "", 400)
		request(t, app, "POST", "/api/lobby/"+lobby.ModelID()+"/join", map[string]string{"code": "room-code"}, guestToken, "", 200)
		// Retrying an accepted join must work even when the room is now full.
		request(t, app, "POST", "/api/lobby/"+lobby.ModelID()+"/join", nil, guestToken, "", 200)
	})
	t.Run("rejoining preserves host role", func(t *testing.T) {
		lobby := room("Host room", false)
		request(t, app, "POST", "/api/lobby/"+lobby.ModelID()+"/join", nil, hostToken, "", 200)
		current, err := repository.GetLobby(lobby.ModelID())
		if err != nil || current.Lobby_Host == nil || current.Lobby_Host.ModelID() != host.ModelID() || len(current.Lobby_Users) != 1 {
			t.Fatalf("rejoining changed ownership/membership: %+v %v", current, err)
		}
	})
	t.Run("outsiders cannot retrieve invitation codes", func(t *testing.T) {
		lobby := room("Hidden", true)
		var one struct {
			Data dbtype.Lobby `json:"data"`
		}
		decode(t, request(t, app, "GET", "/api/lobby/"+lobby.ModelID(), nil, guestToken, "", 200), &one)
		if one.Data.Code != "" {
			t.Fatal("outsider can read private invitation code")
		}
		var list struct {
			Data []dbtype.Lobby `json:"data"`
		}
		decode(t, request(t, app, "GET", "/api/lobby", nil, guestToken, "", 200), &list)
		for _, item := range list.Data {
			if item.Code != "" {
				t.Fatal("list exposes invitation code")
			}
		}
		decode(t, request(t, app, "GET", "/api/lobby/"+lobby.ModelID(), nil, hostToken, "", 200), &one)
		if one.Data.Code != lobby.Code {
			t.Fatal("host cannot retrieve own invitation code")
		}
	})
	t.Run("patch supports false and empty without editing identity", func(t *testing.T) {
		lobby := room("Editable", true)
		request(t, app, "POST", "/api/lobby/"+lobby.ModelID(), map[string]any{"private": false, "code": "", "guid": "tampered", "date_created": "2099-01-01T00:00:00Z"}, hostToken, "", 200)
		current, err := repository.GetLobby(lobby.ModelID())
		if err != nil || current.Private || current.Code != "" || current.Guid != lobby.Guid || current.Date_Created != lobby.Date_Created || current.Name != lobby.Name {
			t.Fatalf("patch lost values or changed identity: %+v %v", current, err)
		}
	})
}
