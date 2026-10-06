//go:build integration

package database_test

import (
	"testing"
	"viral-game-network/src/database"
	"viral-game-network/src/database/migrations"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestDatabaseCRUDAndQueryResults(t *testing.T) {
	support.Storage(t)
	user, err := database.Create(&dbtype.User{Name: "Alice", Guid: database.GetUUID(), Date_Created: "2026-01-01T00:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	selected, err := database.Select[dbtype.User](user.ID.String())
	if err != nil || selected.Name != "Alice" {
		t.Fatalf("%+v %v", selected, err)
	}
	user.Name = "Updated"
	updated, err := database.Update(*user.ID, user)
	if err != nil || updated.Name != "Updated" {
		t.Fatalf("%+v %v", updated, err)
	}
	rows, err := database.Query[dbtype.User]("SELECT * FROM User;", nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("%+v %v", rows, err)
	}
	scalar, err := database.Query[int]("RETURN 7;", nil)
	if err != nil || len(scalar) != 1 || scalar[0] != 7 {
		t.Fatalf("%+v %v", scalar, err)
	}
	only, err := database.Query[dbtype.User]("CREATE ONLY User CONTENT $user;", map[string]interface{}{"user": dbtype.User{Name: "Once", Guid: database.GetUUID(), Date_Created: "2026-01-01T00:00:00Z"}})
	if err != nil || len(only) != 1 {
		t.Fatalf("%+v %v", only, err)
	}
	created, err := database.Query[dbtype.User]("SELECT * FROM User WHERE name='Once';", nil)
	if err != nil || len(created) != 1 {
		t.Fatalf("query executed more than once: %+v %v", created, err)
	}
	if err := database.Delete(*created[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := database.Delete(*user.ID); err != nil {
		t.Fatal(err)
	}
	rows, err = database.Query[dbtype.User]("SELECT * FROM User;", nil)
	if err != nil || len(rows) != 0 {
		t.Fatalf("%+v %v", rows, err)
	}
	if _, err := database.Query[dbtype.User]("INVALID QUERY;", nil); err == nil {
		t.Fatal("invalid query accepted")
	}
}

func TestQueryReportsStatementErrorsAndRollback(t *testing.T) {
	support.Storage(t)
	for _, query := range []string{
		"THROW 'rejected';",
		"BEGIN TRANSACTION; CREATE User CONTENT {name:'Rolled back', guid:'rolled-back', date_created:'2026-01-01T00:00:00Z'}; THROW 'rejected'; COMMIT TRANSACTION;",
	} {
		if _, err := database.Query[any](query, nil); err == nil {
			t.Fatal("failed SQL statement reported success")
		}
	}
	users, err := database.Query[dbtype.User]("SELECT * FROM User;", nil)
	if err != nil || len(users) != 0 {
		t.Fatalf("failed transaction committed: %+v %v", users, err)
	}
}

func TestFailedMigrationIsNotRecorded(t *testing.T) {
	support.Storage(t)
	previous := migrations.AllMigrations
	t.Cleanup(func() { migrations.AllMigrations = previous })
	migrations.AllMigrations = append(append([]dbtype.Migration(nil), previous...), dbtype.Migration{Name: "test-failed-migration", SQL: "DEFINE TABLE TestFailedMigration SCHEMAFULL; THROW 'rejected migration';"})
	if err := migrations.Migrations_Run(database.DBS); err == nil {
		t.Fatal("failed migration reported success")
	}
	rows, err := database.Query[dbtype.Migration]("SELECT * FROM Migrations WHERE name='test-failed-migration';", nil)
	if err != nil || len(rows) != 0 {
		t.Fatalf("failed migration marked applied: %+v %v", rows, err)
	}
	if _, err := database.Query[any]("DEFINE TABLE TestFailedMigration SCHEMAFULL;", nil); err != nil {
		t.Fatalf("failed migration left partial schema: %v", err)
	}
}

func TestSelectMissingRecord(t *testing.T) {
	support.Storage(t)
	user, err := database.Select[dbtype.User]("User:missing")
	if err != nil || user != nil {
		t.Fatalf("missing record: %+v %v", user, err)
	}
}
