package dbtype

import (
	"encoding/json"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"testing"
)

func TestModelIdentity(t *testing.T) {
	for _, model := range []Model{Admin{}, Application{}, Lobby{}, LobbyMessage{}, Migration{}, Server{}, SystemConfig{}, SystemEvent{}, User{}} {
		if model.ModelID() != "" {
			t.Errorf("%T has an ID before persistence", model)
		}
		if model.TableName() == "" {
			t.Errorf("%T has no table", model)
		}
	}
	record := User{ID: &models.RecordID{Table: "User", ID: "alice"}}
	if record.ModelID() != "User:alice" {
		t.Fatal(record.ModelID())
	}
	record.SetGUID("new-guid")
	if record.Guid != "new-guid" {
		t.Fatal("SetGUID did not update the record")
	}
}

func TestApplicationFormRecordID(t *testing.T) {
	var app Application
	if err := app.UnmarshalText([]byte("Application:game")); err != nil || app.ModelID() != "Application:game" {
		t.Fatalf("%+v %v", app, err)
	}
	for _, value := range []string{"broken", "Application:", "User:alice"} {
		if err := app.UnmarshalText([]byte(value)); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if err := app.UnmarshalText(nil); err != nil || app.ID != nil {
		t.Fatal("empty selection did not clear ID")
	}
}

func TestSystemConfigurationMetadata(t *testing.T) {
	config := SystemConfig{ReadOnly: true, SortOrder: 2}
	data, _ := json.Marshal(config)
	var decoded map[string]any
	json.Unmarshal(data, &decoded)
	if decoded["readonly"] != true || decoded["sortorder"] != float64(2) {
		t.Fatalf("metadata collision: %s", data)
	}
	for key, value := range config.KeyValMap() {
		if key != value.Key || value.Type == "" {
			t.Errorf("invalid config %q: %+v", key, value)
		}
	}
}

func TestApplicationJSONRepresentations(t *testing.T) {
	original := Application{ID: &models.RecordID{Table: "Application", ID: "game"}, Name: "Game", Port: "4000"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Application
	if err := json.Unmarshal(data, &decoded); err != nil || decoded.ModelID() != original.ModelID() || decoded.Port != "4000" {
		t.Fatalf("API representation: %s %v", data, err)
	}
	if err := json.Unmarshal([]byte(`"Application:other"`), &decoded); err != nil || decoded.ModelID() != "Application:other" {
		t.Fatalf("form representation: %+v %v", decoded, err)
	}
	if err := json.Unmarshal([]byte(`"User:wrong"`), &decoded); err == nil {
		t.Fatal("wrong table accepted")
	}
}
