//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/cache"
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
	request(t, app, "GET", "/api/host/"+id+"/tick", nil, extraToken, "", 403)
	request(t, app, "GET", "/api/host/"+id+"/tick", nil, memberToken, "", 403)
	serverToken, err := auth.GenerateToken("Game server", id, auth.ServerAudience)
	if err != nil {
		t.Fatal(err)
	}
	request(t, app, "GET", "/api/host/"+id+"/tick", nil, serverToken, "", 200)
	request(t, app, "GET", "/api/host/Lobby:another/tick", nil, serverToken, "", 403)
	request(t, app, "GET", "/api/lobby", nil, serverToken, "", 403)
	request(t, app, "GET", "/admin/users", nil, "", "VNET_SESSION="+serverToken, 302)
	response = request(t, app, "POST", "/api/host/"+id+"/refresh", nil, serverToken, "", 200)
	var refreshed struct {
		Token string `json:"access_token"`
	}
	decode(t, response, &refreshed)
	claims, err := auth.ValidateTokenFor(refreshed.Token, auth.ServerAudience)
	if err != nil || claims.Subject != id {
		t.Fatal("server refresh changed lobby identity", err)
	}
	request(t, app, "POST", "/api/host/"+id+"/refresh", nil, hostToken, "", 403)
	unknown, err := auth.GenerateToken("Deleted User", "User:missing", auth.UserAudience)
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
	if err := bootstrap(configs); err != nil {
		t.Fatal(err)
	}
	if err := bootstrap(configs); err != nil {
		t.Fatal(err)
	}
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
	var sessionResponse map[string]interface{}
	decode(t, response, &sessionResponse)
	if _, exposed := sessionResponse["access_token"]; exposed {
		t.Fatal("admin login exposes its session token to JavaScript")
	}
	if _, exists := sessionResponse["expires_at"]; !exists {
		t.Fatal("admin login has no expiry metadata")
	}
	cookie := ""
	var sessionExpiry int64
	for _, item := range response.Cookies() {
		if item.Name == "VNET_SESSION" {
			cookie = item.Name + "=" + item.Value
			claims, err := auth.ValidateTokenFor(item.Value, auth.AdminAudience)
			if err != nil || !item.Expires.Equal(claims.ExpiresAt.Time) || item.Path != "/" || !item.HttpOnly || item.SameSite != http.SameSiteStrictMode {
				t.Fatalf("invalid session cookie: %v", err)
			}
			sessionExpiry = claims.ExpiresAt.Unix()
			request(t, app, "GET", "/api/lobby", nil, item.Value, "", 403)
		}
	}
	if cookie == "" {
		t.Fatal("session cookie missing")
	}
	for _, path := range []string{"/admin", "/admin/admins", "/admin/applications", "/admin/users", "/admin/lobbies", "/admin/servers", "/admin/configs", "/admin/events", "/admin/metrics", "/admin/admin", "/admin/application", "/admin/user", "/admin/lobby"} {
		response = request(t, app, "GET", path, nil, "", cookie, 200)
		body, err := io.ReadAll(response.Body)
		if !strings.Contains(string(body), `data-authentication-expires-value="`+strconv.FormatInt(sessionExpiry, 10)+`"`) {
			t.Fatalf("view %s missing session expiry metadata", path)
		}
		if err != nil || !strings.Contains(response.Header.Get("Content-Type"), "text/html") || len(body) == 0 {
			t.Fatalf("invalid view %s: %v", path, err)
		}
	}
	refresh := request(t, app, "GET", "/auth/admin/refresh", nil, "", cookie, 200)
	decode(t, refresh, &sessionResponse)
	if _, exposed := sessionResponse["access_token"]; exposed {
		t.Fatal("admin refresh exposes its session token")
	}
	for _, item := range refresh.Cookies() {
		if item.Name == "VNET_SESSION" && !item.HttpOnly {
			t.Fatal("refresh removed HttpOnly")
		}
	}
	response = request(t, app, "GET", "/auth/admin/logout", nil, "", cookie, 302)
	for _, item := range response.Cookies() {
		if item.Name == "VNET_SESSION" && item.Value != "" {
			t.Fatal("logout retained session")
		}
	}
}

func TestBootstrapRejectsIncompleteCredentials(t *testing.T) {
	support.Storage(t)
	for _, config := range []map[string]string{
		{},
		{"adminUsername": "admin", "adminPassword": "password", "appName": "test"},
		{"adminUsername": "admin", "appName": "test", "appKey": "test-key"},
	} {
		if err := bootstrap(config); err == nil {
			t.Fatal("bootstrap accepted incomplete credentials")
		}
	}
	admins, count, err := repository.AllAdmin(10, 1, "")
	if err != nil || count != 0 || len(admins) != 0 {
		t.Fatal("invalid bootstrap created an admin", admins, err)
	}
	keys, err := repository.GetApiKeys()
	if err != nil || len(keys) != 0 {
		t.Fatal("invalid bootstrap created API keys", err)
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
	token, err := auth.GenerateToken(user.Name, user.ModelID(), auth.UserAudience)
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

func TestGuestSessionsCannotImpersonateAdminsOrPlayers(t *testing.T) {
	support.Storage(t)
	t.Setenv("VNET_TOKEN_EXPIRE", "60")
	redis := support.Redis(t)
	t.Setenv("CACHE_ENDPOINT", redis.Addr())
	if err := repository.AddApiKey("test", "test-key"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.PutAdmin(&dbtype.Admin{Name: "admin", Email: "admin@example.invalid", Phone: "123", Password: "unused"}); err != nil {
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
		return result.Token
	}
	t.Run("guest token cannot authorize admin", func(t *testing.T) {
		token := guest("admin")
		request(t, app, "GET", "/admin/users", nil, "", "VNET_SESSION="+token, 302)
		request(t, app, "GET", "/auth/admin", nil, "", "VNET_SESSION="+token, 200)
	})
	t.Run("duplicate display names have distinct identities", func(t *testing.T) {
		first, err := auth.ValidateToken(guest("Player"))
		if err != nil {
			t.Fatal(err)
		}
		second, err := auth.ValidateToken(guest("Player"))
		if err != nil {
			t.Fatal(err)
		}
		if first.Subject == "" || second.Subject == "" || first.Subject == second.Subject {
			t.Fatal("guest sessions share an identity instead of distinct record IDs")
		}
		user, err := repository.GetUser(first.Subject)
		if err != nil {
			t.Fatal(err)
		}
		user.Name = "Renamed player"
		if _, err := repository.SetUser(user.ModelID(), user); err != nil {
			t.Fatal(err)
		}
		original, err := auth.GenerateToken("Player", first.Subject, auth.UserAudience)
		if err != nil {
			t.Fatal(err)
		}
		request(t, app, "GET", "/api/lobby", nil, original, "", 200)
		response := request(t, app, "POST", "/auth/lobby", map[string]string{"name": "admin"}, original, "", 200)
		var renewed struct {
			Token string `json:"access_token"`
		}
		decode(t, response, &renewed)
		claims, err := auth.ValidateTokenFor(renewed.Token, auth.UserAudience)
		if err != nil || claims.Subject != first.Subject || claims.Username != user.Name {
			t.Fatalf("refresh changed identity: %+v %v", claims, err)
		}
		if err := repository.DelUser(user.ModelID(), user); err != nil {
			t.Fatal(err)
		}
		request(t, app, "GET", "/api/lobby", nil, renewed.Token, "", 403)
	})
	t.Run("lobby login requires an authenticated session", func(t *testing.T) {
		request(t, app, "POST", "/auth/lobby", map[string]string{"name": "Player"}, "", "", 403)
	})
}

func TestLobbySocketRevokesDepartedMembership(t *testing.T) {
	for _, action := range []string{"receive", "send", "quiet"} {
		t.Run(action, func(t *testing.T) {
			redis := support.Storage(t)
			t.Setenv("CACHE_ENDPOINT", redis.Addr())
			user, err := repository.PutUser(&dbtype.User{Name: "Host"})
			if err != nil {
				t.Fatal(err)
			}
			game, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Port: "4000", Command: "server", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
			if err != nil {
				t.Fatal(err)
			}
			room, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, game, user)
			if err != nil {
				t.Fatal(err)
			}
			token, err := auth.GenerateToken(user.Name, user.ModelID(), auth.UserAudience)
			if err != nil {
				t.Fatal(err)
			}
			if err := ministration.Service_Lobby_Notify(room.ModelID(), "welcome", ""); err != nil {
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
			connection, _, err := websocket.DefaultDialer.Dial("ws://"+listener.Addr().String()+"/api/lobby/"+room.ModelID()+"/socket", http.Header{"Viral-Game-Network-Token": []string{token}})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close()
			connection.SetReadDeadline(time.Now().Add(3 * time.Second))
			if _, _, err := connection.ReadMessage(); err != nil {
				t.Fatal("socket did not finish replay", err)
			}
			if err := repository.UnlinkLobbyUser(room, user); err != nil {
				t.Fatal(err)
			}
			switch action {
			case "receive":
				if err := ministration.Service_Lobby_Notify(room.ModelID(), "private message", ""); err != nil {
					t.Fatal(err)
				}
			case "send":
				if err := connection.WriteMessage(websocket.TextMessage, []byte("unauthorized chat")); err != nil {
					t.Fatal(err)
				}
			}
			_, raw, err := connection.ReadMessage()
			var timeout net.Error
			if err == nil || (errors.As(err, &timeout) && timeout.Timeout()) {
				t.Fatalf("departed member retained chat access: payload=%s error=%v", raw, err)
			}
			if action == "send" {
				history, err := cache.List("lobby:" + room.ModelID() + ":channel")
				if err != nil || len(history) != 1 {
					t.Fatalf("departed member published to history: %v %v", history, err)
				}
			}
		})
	}
}
