package form_builder

import (
	"encoding/json"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"testing"
	dbtype "viral-game-network/src/database/type"
)

func TestAdminFormModes(t *testing.T) {
	for _, mode := range []string{"create", "update"} {
		t.Run(mode, func(t *testing.T) {
			form, err := GenerateForm("POST", "/admin/admin", mode, dbtype.Admin{Name: "Alice"}, "Admin", "")
			if err != nil {
				t.Fatal(err)
			}
			if form.Submit != "Submit" || form.Action != "/admin/admin" || !form.CanReset || form.DisableSubmit {
				t.Fatalf("bad form: %+v", form)
			}
			fields := map[string]FormField{}
			for _, field := range form.Fields {
				fields[field.Name] = field
			}
			if fields["name"].Value != "Alice" || !fields["name"].Required || !fields["id"].ReadOnly {
				t.Fatal("field metadata lost")
			}
			if fields["password"].Required != (mode == "create") {
				t.Fatal("password requirement does not respect mode")
			}
			if _, ok := fields["date_created"]; ok {
				t.Fatal("non-form field exposed")
			}
		})
	}
}

func TestFormValuesAndMetadata(t *testing.T) {
	type input struct {
		Application *dbtype.Application `json:"application,omitempty" form:"label:App,type:choice"`
		Enabled     bool                `json:"enabled" form:"label:Enabled,type:bool,sortorder:7,readonly-update:true"`
		Secret      string              `json:"-" form:"label:Secret,type:text"`
	}
	record := input{Application: &dbtype.Application{ID: &models.RecordID{Table: "Application", ID: "game"}}, Enabled: true}
	form, err := GenerateForm("POST", "/", "update", &record, "", "Save")
	if err != nil {
		t.Fatal(err)
	}
	if len(form.Fields) != 2 || form.Fields[0].Value != "Application:game" || form.Fields[1].Value != "true" || !form.Fields[1].ReadOnly {
		t.Fatalf("unexpected fields: %+v", form.Fields)
	}
	data, _ := json.Marshal(form.Fields[1])
	var decoded map[string]any
	json.Unmarshal(data, &decoded)
	if decoded["readonly"] != true || decoded["sortorder"] != float64(7) {
		t.Fatalf("metadata collision: %s", data)
	}
}

func TestFormRejectsInvalidInputs(t *testing.T) {
	var pointer *dbtype.Admin
	for _, input := range []any{nil, pointer, "string", 42, []string{}} {
		if _, err := GenerateForm("POST", "/", "create", input, "", ""); err == nil {
			t.Errorf("accepted %T", input)
		}
	}
}
