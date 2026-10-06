package job

import (
	"github.com/robfig/cron/v3"
	"testing"
)

func TestJobRegistry(t *testing.T) {
	expected := map[string]bool{"monitoring": true, "lobby_server_provisioner": true, "lobby_server_stewardship": true, "lobby_stewardship": true}
	seen := map[string]bool{}
	scheduler := cron.New(cron.WithSeconds())
	for _, entry := range Stack() {
		if !expected[entry.Name] || seen[entry.Name] || entry.Func == nil {
			t.Fatalf("invalid job registration %+v", entry)
		}
		seen[entry.Name] = true
		if _, err := scheduler.AddFunc(entry.Time, entry.Func); err != nil {
			t.Fatalf("invalid schedule for %s: %v", entry.Name, err)
		}
	}
	if len(seen) != len(expected) {
		t.Fatal("missing job")
	}
}
