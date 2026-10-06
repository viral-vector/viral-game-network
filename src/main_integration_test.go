//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/ministration"
	"viral-game-network/tests/support"
)

func request(t *testing.T, app *fiber.App, method, path string, payload any, token, cookie string, status int) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Viral-Game-Network-Token", token)
	req.Header.Set("Viral-Game-Network-AppKey", "test-key")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	response, err := app.Test(req, 60000)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		t.Fatalf("%s %s: want %d, got %d: %s", method, path, status, response.StatusCode, data)
	}
	t.Cleanup(func() { response.Body.Close() })
	return response
}

func decode(t *testing.T, response *http.Response, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func TestGuestLobbyAndHeartbeatFlow(t *testing.T) {
	support.Storage(t)
	t.Setenv("VNET_TOKEN_EXPIRE", "60")
	redis := support.Redis(t)
	t.Setenv("CACHE_ENDPOINT", redis.Addr())
	if err := repository.AddApiKey("test", "test-key"); err != nil {
		t.Fatal(err)
	}
	app := NewApp("../views", "../public")
	t.Cleanup(func() { app.Shutdown() })
	guest := func(name string) string {
		response := request(t, app, "POST", "/auth/guest", map[string]string{"name": name}, "", "", 200)
		var result struct {
			Token string `json:"access_token"`
		}
		decode(t, response, &result)
		if result.Token == "" {
			t.Fatal("guest token missing")
		}
		return result.Token
	}
	hostToken := guest("Host")
	memberToken := guest("Member")
	extraToken := guest("Extra")
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "node index.js", Lobby_Max_Players: "2", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	response := request(t, app, "POST", "/api/lobby/host", map[string]string{"name": "Room", "app": application.Guid}, hostToken, "", 200)
	var hosted struct {
		Data dbtype.Lobby `json:"data"`
	}
	decode(t, response, &hosted)
	id := hosted.Data.ModelID()
	if id == "" || hosted.Data.Lobby_Host == nil {
		t.Fatal("host relationship missing")
	}
	request(t, app, "POST", "/api/lobby/"+id, map[string]string{"name": "Unauthorized"}, memberToken, "", 403)
	request(t, app, "POST", "/api/lobby/"+id, map[string]string{"name": "Updated"}, hostToken, "", 200)
	response = request(t, app, "GET", "/api/lobby/"+id, nil, hostToken, "", 200)
	decode(t, response, &hosted)
	if hosted.Data.Name != "Updated" {
		t.Fatal("lobby update not persisted")
	}
	request(t, app, "GET", "/api/host/"+id+"/tick", nil, hostToken, "", 400)
	request(t, app, "POST", "/api/lobby/"+id+"/join", nil, memberToken, "", 200)
	request(t, app, "POST", "/api/lobby/"+id+"/join", nil, extraToken, "", 400)
	request(t, app, "POST", "/api/lobby/"+id, map[string]string{"name": "Unauthorized"}, memberToken, "", 403)
	server, err := repository.PutServer(&dbtype.Server{Name: "Server", Guid: "server", Status: "Running", Address: "203.0.113.1", Port: 30000}, &hosted.Data)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.LinkLobbyServer(&hosted.Data, server); err != nil {
		t.Fatal(err)
	}
	response = request(t, app, "GET", "/api/host/"+id+"/tick", nil, hostToken, "", 200)
	var heartbeat struct {
		Data dbtype.Server `json:"data"`
	}
	decode(t, response, &heartbeat)
	if heartbeat.Data.Status != "Online" {
		t.Fatal("heartbeat did not mark server online")
	}
	unknown, err := auth.GenerateToken("Deleted User")
	if err != nil {
		t.Fatal(err)
	}
	request(t, app, "GET", "/api/lobby", nil, unknown, "", 403)
	request(t, app, "POST", "/api/lobby/Lobby:missing/join", nil, extraToken, "", 400)
}

func TestBootstrapAndAdminSessionViews(t *testing.T) {
	support.Storage(t)
	t.Setenv("VNET_TOKEN_EXPIRE", "60")
	redis := support.Redis(t)
	t.Setenv("CACHE_ENDPOINT", redis.Addr())
	configs := map[string]string{"adminUsername": "admin", "adminPassword": "password", "appName": "test", "appKey": "test-key"}
	bootstrap(configs)
	bootstrap(configs)
	if admins, total, err := repository.AllAdmin(10, 1, ""); err != nil || total != 1 || len(admins) != 1 {
		t.Fatalf("bootstrap duplicated admins: %+v %d %v", admins, total, err)
	}
	if keys, err := repository.GetApiKeys(); err != nil || len(keys) != 1 {
		t.Fatalf("bootstrap duplicated keys: %+v %v", keys, err)
	}
	app := NewApp("../views", "../public")
	t.Cleanup(func() { app.Shutdown() })
	request(t, app, "GET", "/auth/admin", nil, "", "", 200)
	request(t, app, "POST", "/auth/admin", map[string]string{"username": "admin", "password": "wrong"}, "", "", 400)
	response := request(t, app, "POST", "/auth/admin", map[string]string{"username": "admin", "password": "password"}, "", "", 200)
	cookie := ""
	for _, item := range response.Cookies() {
		if item.Name == "VNET_SESSION" {
			cookie = item.Name + "=" + item.Value
		}
	}
	if cookie == "" {
		t.Fatal("session cookie missing")
	}
	for _, path := range []string{"/admin", "/admin/admins", "/admin/applications", "/admin/users", "/admin/lobbies", "/admin/servers", "/admin/configs", "/admin/events", "/admin/metrics", "/admin/admin", "/admin/application", "/admin/user", "/admin/lobby"} {
		response = request(t, app, "GET", path, nil, "", cookie, 200)
		body, err := io.ReadAll(response.Body)
		if err != nil || !strings.Contains(response.Header.Get("Content-Type"), "text/html") || len(body) == 0 {
			t.Fatalf("invalid view %s: %v", path, err)
		}
	}
	request(t, app, "GET", "/auth/admin/refresh", nil, "", cookie, 200)
	response = request(t, app, "GET", "/auth/admin/logout", nil, "", cookie, 302)
	for _, item := range response.Cookies() {
		if item.Name == "VNET_SESSION" && item.Value != "" {
			t.Fatal("logout retained session")
		}
	}
}

func TestLobbySocketReplaysHistoryAndBroadcastsMessages(t *testing.T) {
	redis := support.Storage(t)
	t.Setenv("CACHE_ENDPOINT", redis.Addr())
	t.Setenv("VNET_TOKEN_EXPIRE", "60")
	user, err := repository.PutUser(&dbtype.User{Name: "Host"})
	if err != nil {
		t.Fatal(err)
	}
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "node index.js", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, application, user)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.GenerateToken(user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := ministration.Service_Lobby_Notify(lobby.ModelID(), "Earlier message", user.ModelID()); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	app := NewApp("../views", "../public")
	appDone := make(chan error, 1)
	go func() { appDone <- app.Listener(listener) }()
	t.Cleanup(func() {
		app.Shutdown()
		if err := <-appDone; err != nil {
			t.Error(err)
		}
	})
	url := "ws://" + listener.Addr().String() + "/api/lobby/" + lobby.ModelID() + "/socket"
	connection, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Viral-Game-Network-Token": []string{token}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	connection.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, message, err := connection.ReadMessage()
	if err != nil {
		t.Fatal("history was not replayed", err)
	}
	var history dbtype.LobbyMessage
	if err := json.Unmarshal(message, &history); err != nil || history.Body != "Earlier message" {
		t.Fatalf("history: %s %v", message, err)
	}
	if err := connection.WriteMessage(websocket.TextMessage, []byte("Hi")); err != nil {
		t.Fatal(err)
	}
	_, message, err = connection.ReadMessage()
	if err != nil || !bytes.Contains(message, []byte("Message length")) {
		t.Fatalf("short message accepted: %s %v", message, err)
	}
	if err := connection.WriteMessage(websocket.TextMessage, []byte("New message")); err != nil {
		t.Fatal(err)
	}
	_, message, err = connection.ReadMessage()
	if err != nil {
		t.Fatal("broadcast was not received", err)
	}
	if err := json.Unmarshal(message, &history); err != nil || history.Body != "New message" || history.User_ID != user.ModelID() {
		t.Fatalf("broadcast: %s %v", message, err)
	}
}
