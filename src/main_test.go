package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"viral-game-network/tests/support"

	"github.com/gofiber/fiber/v3"
)

func TestPublicRoutesAndAuthenticationBoundary(t *testing.T) {
	server := support.Redis(t)
	t.Setenv("CACHE_ENDPOINT", server.Addr())
	app := NewApp("../views", "../public")
	defer app.Shutdown()
	for _, tc := range []struct {
		path   string
		status int
	}{{"/", 200}, {"/health", 200}, {"/api/lobby", 403}, {"/admin", 302}, {"/missing", 404}} {
		t.Run(tc.path, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.status {
				body, _ := io.ReadAll(response.Body)
				t.Fatalf("status %d: %s", response.StatusCode, body)
			}
			if tc.path == "/health" {
				var body map[string]string
				json.NewDecoder(response.Body).Decode(&body)
				if body["status"] != "success" {
					t.Fatal(body)
				}
			}
		})
	}
}

func TestRouteMetadataPreservesSearchAndPagingQuery(t *testing.T) {
	server := support.Redis(t)
	t.Setenv("CACHE_ENDPOINT", server.Addr())
	app := NewApp("../views", "../public")
	defer app.Shutdown()
	app.Get("/inspect-route", func(c fiber.Ctx) error { return c.JSON(c.Locals("route")) })
	path := "/inspect-route?search=some%20game&page=2&pageToken=cursor%2Bvalue"
	for _, target := range []string{path, "http://example.com" + path} {
		response, err := app.Test(httptest.NewRequest("GET", target, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var route map[string]string
		if err := json.NewDecoder(response.Body).Decode(&route); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 200 || route["path"] != "/inspect-route" || route["http"] != "http://example.com"+path {
			t.Fatalf("target=%s status=%d route=%v", target, response.StatusCode, route)
		}
	}
}

func TestBootstrapReturnsStorageFailure(t *testing.T) {
	support.Redis(t)
	if err := bootstrap(map[string]string{}); err == nil {
		t.Fatal("bootstrap accepted missing storage")
	}
}
