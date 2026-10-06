//go:build integration

package handler_admin

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"viral-game-network/src/database"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func crudRequest(t *testing.T, handler fiber.Handler, method, id string, body any, status int) {
	t.Helper()
	app := fiber.New()
	app.Use(recover.New())
	app.Add(method, "/:id", handler)
	var encoded bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, "/"+id, &encoded)
	req.Header.Set("Content-Type", "application/json")
	response, err := app.Test(req, 60000)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != status {
		text, _ := io.ReadAll(response.Body)
		t.Fatalf("%s %s: want %d, got %d: %s", method, id, status, response.StatusCode, text)
	}
}

func TestPartialUserUpdatePreservesIdentity(t *testing.T) {
	support.Storage(t)
	user, err := repository.PutUser(&dbtype.User{Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	crudRequest(t, Handle_Users_Update_Crud, "POST", user.ModelID(), map[string]string{"name": "Renamed"}, 200)
	actual, err := repository.GetUser(user.ModelID())
	if err != nil || actual.Name != "Renamed" || actual.Guid != user.Guid || actual.Date_Created != user.Date_Created {
		t.Fatalf("partial update lost user fields: %+v %v", actual, err)
	}
}

func TestPartialApplicationUpdatePreservesConfiguration(t *testing.T) {
	support.Storage(t)
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "image", Version: "1", Port: "4000", Command: "run game", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	crudRequest(t, Handle_Applications_Update_Crud, "POST", application.ModelID(), map[string]string{"name": "Renamed"}, 200)
	actual, err := repository.GetApplication(application.ModelID())
	if err != nil || actual.Name != "Renamed" || actual.Guid != application.Guid || actual.Image != application.Image || actual.Port != application.Port || actual.Command != application.Command {
		t.Fatalf("partial update lost application fields: %+v %v", actual, err)
	}
}

func TestPartialAdminUpdatePreservesCredentials(t *testing.T) {
	support.Storage(t)
	admin, err := repository.PutAdmin(&dbtype.Admin{Name: "Admin", Email: "admin@example.invalid", Phone: "123", Password: "preserved-hash"})
	if err != nil {
		t.Fatal(err)
	}
	crudRequest(t, Handle_Admins_Update_Crud, "POST", admin.ModelID(), map[string]string{"name": "Renamed"}, 200)
	actual, err := repository.GetAdmin(admin.ModelID())
	if err != nil || actual.Name != "Renamed" || actual.Password != admin.Password || actual.Email != admin.Email || actual.Phone != admin.Phone {
		t.Fatalf("partial update lost admin fields: %+v %v", actual, err)
	}
}

func TestAdminCRUDRejectsMissingLobbyInputsAndRecords(t *testing.T) {
	support.Storage(t)
	crudRequest(t, Handle_Lobbies_Create_Crud, "POST", "new", map[string]string{"name": "Room"}, 400)
	crudRequest(t, Handle_Lobbies_Update_Crud, "POST", "Lobby:missing", map[string]string{"name": "Room"}, 404)
	crudRequest(t, Handle_Lobbies_Delete_Crud, "DELETE", "Lobby:missing", nil, 404)
	crudRequest(t, Handle_Servers_Delete_Crud, "DELETE", "Server:missing", nil, 404)
}

func TestApplicationDeletionWorksAndProtectsActiveLobbies(t *testing.T) {
	support.Storage(t)
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "image", Version: "1", Port: "4000", Command: "run game", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	crudRequest(t, Handle_Applications_Delete_Crud, "DELETE", application.ModelID(), nil, 200)
	if actual, err := repository.GetApplication(application.ModelID()); err == nil || actual != nil {
		t.Fatal("application deletion reported success but retained the record")
	}
	application, err = repository.PutApplication(&dbtype.Application{Name: "Active", Guid: "active", Image: "image", Version: "1", Port: "4000", Command: "run game", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, application, nil)
	if err != nil {
		t.Fatal(err)
	}
	crudRequest(t, Handle_Applications_Delete_Crud, "DELETE", application.ModelID(), nil, 409)
	if current, err := repository.GetLobby(lobby.ModelID()); err != nil || current.Lobby_Application == nil || current.Lobby_Application.ID == nil {
		t.Fatal("deleting an in-use application broke its lobby", err)
	}
}

func TestSettingsHandlerRejectsFailedReadsAndInvalidExpiration(t *testing.T) {
	support.Storage(t)
	configs, err := repository.GetSystemConfigs()
	if err != nil {
		t.Fatal(err)
	}
	for i := range configs {
		if configs[i].Key == "VNET_NAME" {
			configs[i].Val = "Original"
		} else if configs[i].Key == "VNET_TOKEN_EXPIRE" {
			configs[i].Val = "45"
		}
	}
	if _, err := repository.PopSystemConfigs(&configs); err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/", Handle_Configs_Update_Crud)
	for _, body := range []string{"VNET_TOKEN_EXPIRE=-1", "VNET_TOKEN_EXPIRE=not-a-number", "VNET_TOKEN_EXPIRE=0"} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Fatalf("invalid expiration accepted: %d", response.StatusCode)
		}
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader("VNET_NAME=Renamed"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("partial settings update failed: %d", response.StatusCode)
	}
	lifetime, err := repository.SelConfig("VNET_TOKEN_EXPIRE")
	if err != nil || lifetime.Val != "45" {
		t.Fatalf("partial settings update erased token lifetime: %+v %v", lifetime, err)
	}
	repository.DelSystemConfigsCache()
	database.Close()
	req = httptest.NewRequest("POST", "/", strings.NewReader("VNET_NAME=Another"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 500 {
		t.Fatalf("failed database read reported success: %d", response.StatusCode)
	}
}
