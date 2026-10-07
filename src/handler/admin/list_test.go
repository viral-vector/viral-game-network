package handler_admin

import (
	"io"
	"net/http/httptest"
	"testing"
	"viral-game-network/src/database"

	"github.com/gofiber/fiber/v3"
)

type listViews struct{ calls int }

func (*listViews) Load() error { return nil }
func (v *listViews) Render(w io.Writer, _ string, _ interface{}, _ ...string) error {
	v.calls++
	_, err := io.WriteString(w, "list rendered")
	return err
}

func TestAdminListsRejectBadPagesAndReportStoreFailures(t *testing.T) {
	previous := database.DBS
	database.DBS = nil
	t.Cleanup(func() { database.DBS = previous })
	for name, handler := range map[string]fiber.Handler{
		"users": Handle_Users, "admins": Handle_Admins, "applications": Handle_Applications,
		"lobbies": Handle_Lobbies, "servers": Handle_Servers, "events": Handle_Events, "pods": Handle_Pods,
	} {
		t.Run(name, func(t *testing.T) {
			views := &listViews{}
			app := fiber.New(fiber.Config{Views: views})
			app.Get("/", handler)
			for _, input := range []struct {
				page   string
				status int
			}{
				{"0", 400}, {"-1", 400}, {"abc", 400}, {"9223372036854775807", 400}, {"1", 500},
			} {
				response, err := app.Test(httptest.NewRequest("GET", "/?page="+input.page, nil))
				if err != nil {
					t.Fatal(err)
				}
				response.Body.Close()
				if response.StatusCode != input.status {
					t.Errorf("page=%s: got %d; want %d", input.page, response.StatusCode, input.status)
				}
			}
			if views.calls != 0 {
				t.Fatalf("invalid or failed list rendered %d times", views.calls)
			}
		})
	}
}

func TestPodListRejectsInvalidPreviousPage(t *testing.T) {
	views := &listViews{}
	app := fiber.New(fiber.Config{Views: views})
	app.Get("/", Handle_Pods)
	for _, page := range []string{"0", "-1", "abc", "9223372036854775807"} {
		response, err := app.Test(httptest.NewRequest("GET", "/?prevPage="+page, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Errorf("previous page %s: got %d; want 400", page, response.StatusCode)
		}
	}
	if views.calls != 0 {
		t.Fatal("invalid list rendered")
	}
}
