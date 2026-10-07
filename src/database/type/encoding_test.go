package dbtype

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

func TestDatabaseEncodingExcludesProjectionsButRetainsAPIFields(t *testing.T) {
	user := &User{Name: "Host"}
	lobby := Lobby{
		ID: &models.RecordID{Table: "Lobby", ID: "room"}, Name: "Room",
		Lobby_Application: &Application{Name: "Game"}, Lobby_Host: user,
		Lobby_Users: []*User{user}, Lobby_Server: &Server{Name: "Server"},
	}
	server := Server{ID: &models.RecordID{Table: "Server", ID: "server"}, Name: "Server", Lobby: &Lobby{Name: "Room"}}
	for _, input := range []struct {
		value  any
		fields []string
	}{
		{lobby, []string{"lobby_application", "lobby_host", "lobby_users", "lobby_server"}},
		{server, []string{"lobby"}},
		{SystemConfig{ID: &models.RecordID{Table: "System_Config", ID: "setting"}, Key: "VNET_NAME", Val: "Network", Name: "Label", Options: []map[string]string{{"label": "Network"}}, ReadOnly: true, SortOrder: 1}, []string{"name", "options", "readonly", "sortorder"}},
	} {
		encoded, err := cbor.Marshal(input.value)
		if err != nil {
			t.Fatal(err)
		}
		var stored map[string]cbor.RawMessage
		if err := cbor.Unmarshal(encoded, &stored); err != nil {
			t.Fatal(err)
		}
		api, err := json.Marshal(input.value)
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]json.RawMessage
		if err := json.Unmarshal(api, &response); err != nil {
			t.Fatal(err)
		}
		for _, field := range input.fields {
			if _, exists := stored[field]; exists {
				t.Errorf("projection %s persisted", field)
			}
			if _, exists := response[field]; !exists {
				t.Errorf("projection %s removed from API", field)
			}
		}
		field := "name"
		if _, ok := input.value.(SystemConfig); ok {
			field = "val"
		}
		if stored[field] == nil || stored["id"] == nil {
			t.Fatal("stored fields lost")
		}
	}
}

func TestDatabaseDecodingRetainsGraphProjections(t *testing.T) {
	encoded, err := cbor.Marshal(map[string]any{
		"id": &models.RecordID{Table: "Lobby", ID: "room"}, "name": "Room",
		"lobby_application": &Application{Name: "Game"}, "lobby_host": &User{Name: "Host"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var lobby Lobby
	if err := cbor.Unmarshal(encoded, &lobby); err != nil {
		t.Fatal(err)
	}
	if lobby.ModelID() != "Lobby:room" || lobby.Lobby_Application == nil || lobby.Lobby_Application.Name != "Game" || lobby.Lobby_Host == nil {
		t.Fatalf("lost graph projections: %+v", lobby)
	}
}
