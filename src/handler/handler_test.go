package handler

import (
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRejectMalformedRequestBodies(t *testing.T) {
	for _, handler := range []fiber.Handler{Handle_AuthLobby, Handle_AuthGuest, Handle_HostLobby, Handle_SetLobby} {
		app := fiber.New()
		app.Post("/", handler)
		request := httptest.NewRequest("POST", "/", strings.NewReader("{broken"))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode < 400 {
			t.Fatalf("malformed body accepted: %d", response.StatusCode)
		}
	}
}

func TestRejectMissingAndMalformedUserTokens(t *testing.T) {
	app := fiber.New()
	app.Use(Handle_ValidateTokenUsers)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	for _, token := range []string{"", "invalid"} {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Set("Viral-Game-Network-Token", token)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 403 {
			t.Fatalf("invalid token accepted: %d", response.StatusCode)
		}
	}
}
