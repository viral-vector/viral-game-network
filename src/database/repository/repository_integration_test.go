//go:build integration

package repository

import (
	"testing"
	"viral-game-network/src/database"
	"viral-game-network/src/database/migrations"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestRepositoryLifecycle(t *testing.T) {
	support.Storage(t)
	user, err := PutUser(&dbtype.User{Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if user.Guid == "" || user.Date_Created == "" {
		t.Fatal("user metadata missing")
	}
	if selected, err := SelUser(&dbtype.User{Guid: user.Guid}); err != nil || selected.ModelID() != user.ModelID() {
		t.Fatalf("%+v %v", selected, err)
	}
	user.Name = "Updated Alice"
	if _, err := SetUser(user.ModelID(), user); err != nil {
		t.Fatal(err)
	}
	if rows, total, err := AllUser(1, 1, ""); err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("%+v %d %v", rows, total, err)
	}
	app, err := PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "server", Lobby_Max_Players: "2", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	if selected, err := SelApplication(&dbtype.Application{Guid: "game"}); err != nil || selected.ModelID() != app.ModelID() {
		t.Fatalf("%+v %v", selected, err)
	}
	app.Name = "Updated Game"
	if _, err := SetApplication(app.ModelID(), app); err != nil {
		t.Fatal(err)
	}
	if rows, total, err := AllApplication(10, 1, ""); err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("%+v %d %v", rows, total, err)
	}
	lobby, err := PutLobby(&dbtype.Lobby{Name: "Room"}, app, user)
	if err != nil {
		t.Fatal(err)
	}
	if lobby.Lobby_Application == nil || lobby.Lobby_Host == nil || lobby.Lobby_Host.ModelID() != user.ModelID() {
		t.Fatalf("missing relationships: %+v", lobby)
	}
	guest, err := PutUser(&dbtype.User{Name: "Guest"})
	if err != nil {
		t.Fatal(err)
	}
	if err := LinkLobbyUser(lobby, guest, "user"); err != nil {
		t.Fatal(err)
	}
	lobby, err = GetLobby(lobby.ModelID())
	if err != nil || len(lobby.Lobby_Users) != 2 {
		t.Fatalf("%+v %v", lobby, err)
	}
	second, err := PutLobby(&dbtype.Lobby{Name: "Second Room"}, app, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := LinkLobbyUser(second, guest, "user"); err != nil {
		t.Fatal(err)
	}
	lobby, err = GetLobby(lobby.ModelID())
	if err != nil || len(lobby.Lobby_Users) != 1 {
		t.Fatalf("user did not leave previous lobby: %+v %v", lobby, err)
	}
	if err := UnlinkLobbyUser(second, guest); err != nil {
		t.Fatal(err)
	}
	second, err = GetLobby(second.ModelID())
	if err != nil || len(second.Lobby_Users) != 0 {
		t.Fatalf("user was not unlinked: %+v %v", second, err)
	}
	server, err := PutServer(&dbtype.Server{Name: "Game", Guid: "server", Status: "Pending", Address: "203.0.113.1", Port: 30000}, lobby)
	if err != nil {
		t.Fatal(err)
	}
	if err := LinkLobbyServer(lobby, server); err != nil {
		t.Fatal(err)
	}
	lobby, err = GetLobby(lobby.ModelID())
	if err != nil || lobby.Lobby_Server == nil || lobby.Lobby_Server.Port != 30000 || lobby.Lobby_Server.Status != "Pending" {
		t.Fatalf("server fields missing: %+v %v", lobby.Lobby_Server, err)
	}
	if ticked, err := TickServer(server.ModelID()); err != nil || ticked.Status != "Online" {
		t.Fatalf("%+v %v", ticked, err)
	}
	if rows, total, err := AllServer(10, 1, ""); err != nil || total != 1 || len(rows) != 1 || rows[0].Lobby == nil {
		t.Fatalf("%+v %d %v", rows, total, err)
	}
	if ports := GetServerPorts(); len(ports) != 1 || ports[0] != 30000 {
		t.Fatal(ports)
	}
	if _, err := GetServerByGuid("server"); err != nil {
		t.Fatal(err)
	}
	if rows, _, err := AllLobbyNotRunning(-1, 1); err != nil || len(rows) != 1 {
		t.Fatalf("%+v %v", rows, err)
	}
	if err := DelServer(server.ModelID(), server); err != nil {
		t.Fatal(err)
	}
	if err := DelLobby(lobby.ModelID(), lobby); err != nil {
		t.Fatal(err)
	}
	if _, err := GetUser(user.ModelID()); err != nil {
		t.Fatal(err)
	}
	if err := DelUser(guest.ModelID(), guest); err != nil {
		t.Fatal(err)
	}
}

func TestConfigurationAdminsAndEvents(t *testing.T) {
	support.Storage(t)
	if rows, total, err := AllSystemEvents(10, 1); err != nil || total != 0 || len(rows) != 0 {
		t.Fatalf("empty events: %+v %d %v", rows, total, err)
	}
	admin, err := GenAdmin("admin", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if selected, err := SelAdmin(&dbtype.Admin{Name: "admin"}); err != nil || selected.ModelID() != admin.ModelID() {
		t.Fatalf("%+v %v", selected, err)
	}
	admin.Phone = "updated-phone"
	if _, err := SetAdmin(admin.ModelID(), admin); err != nil {
		t.Fatal(err)
	}
	if _, err := GetAdmin(admin.ModelID()); err != nil {
		t.Fatal(err)
	}
	if rows, total, err := AllAdmin(10, 1, ""); err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("%+v %d %v", rows, total, err)
	}
	if err := AddApiKey("game", "secret"); err != nil {
		t.Fatal(err)
	}
	keys, err := GetApiKeys()
	if err != nil || len(keys) != 1 || keys[0].Val != "secret" {
		t.Fatalf("%+v %v", keys, err)
	}
	configs, err := GetSystemConfigs()
	if err != nil {
		t.Fatal(err)
	}
	for i := range configs {
		if configs[i].Key == "VNET_NAME" {
			configs[i].Val = "Test"
		}
	}
	if _, err := PopSystemConfigs(&configs); err != nil {
		t.Fatal(err)
	}
	if name := GetConfigValue("VNET_NAME"); name == nil || *name != "Test" {
		t.Fatal(name)
	}
	event, err := PutSystemEvent(&dbtype.SystemEvent{Message: "Created", Severity: "info"})
	if err != nil || event.Date_Created == "" {
		t.Fatalf("%+v %v", event, err)
	}
	if events, err := SelSystemEventsInFrame(60); err != nil || len(events) != 1 {
		t.Fatalf("%+v %v", events, err)
	}
	if events, err := SelSystemEvents(1); err != nil || len(events) != 1 {
		t.Fatalf("%+v %v", events, err)
	}
	if err := migrations.Migrations_Run(database.DBS); err != nil {
		t.Fatal("migrations are not idempotent", err)
	}
	if rows, err := database.Query[dbtype.Migration]("SELECT * FROM Migrations;", nil); err != nil || len(rows) != len(migrations.AllMigrations) {
		t.Fatalf("migration history: %+v %v", rows, err)
	}
	if err := DelAdmin(admin.ModelID(), admin); err != nil {
		t.Fatal(err)
	}
	config, err := SelConfig("API_KEY_game")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := GetConfig(config.ModelID()); err != nil {
		t.Fatal(err)
	}
	if err := DelConfig(config.ModelID(), config); err != nil {
		t.Fatal(err)
	}
}

func TestMissingServerHeartbeatReturnsError(t *testing.T) {
	support.Storage(t)
	if server, err := TickServer("Server:missing"); err == nil || server != nil {
		t.Fatalf("missing server heartbeat: %+v %v", server, err)
	}
}
