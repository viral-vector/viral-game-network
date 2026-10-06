package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"viral-game-network/tests/support"
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
