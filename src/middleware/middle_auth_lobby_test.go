package middleware

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"testing"
	dbtype "viral-game-network/src/database/type"
)

func TestLobbyMembershipUsesRecordValues(t *testing.T) {
	user := func(id string) *dbtype.User { return &dbtype.User{ID: &models.RecordID{Table: "User", ID: id}} }
	lobby := &dbtype.Lobby{Lobby_Host: user("host"), Lobby_Users: []*dbtype.User{user("member"), nil}}
	for _, tc := range []struct {
		name    string
		user    *dbtype.User
		allowed bool
	}{{"host", user("host"), true}, {"member", user("member"), true}, {"outsider", user("other"), false}, {"missing", nil, false}, {"unpersisted", &dbtype.User{}, false}} {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasLobbyAccess(lobby, tc.user); got != tc.allowed {
				t.Fatalf("access=%v, want %v", got, tc.allowed)
			}
		})
	}
	if hasLobbyAccess(nil, user("host")) {
		t.Fatal("missing lobby accepted")
	}
}
