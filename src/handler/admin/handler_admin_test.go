package handler_admin

import (
	"github.com/gofiber/fiber/v3"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateHandlersRejectMalformedBodies(t *testing.T) {
	for _, handler := range []fiber.Handler{Handle_Admins_Create_Crud, Handle_Applications_Create_Crud, Handle_Users_Create_Crud, Handle_Lobbies_Create_Crud, Handle_Configs_ApiKey_Create_Crud} {
		app := fiber.New()
		app.Post("/", handler)
		request := httptest.NewRequest("POST", "/", strings.NewReader("{broken"))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Fatalf("status=%d, want 400", response.StatusCode)
		}
	}
}
