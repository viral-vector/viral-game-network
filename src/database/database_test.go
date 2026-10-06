package database

import (
	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"testing"
	dbtype "viral-game-network/src/database/type"
)

func TestQueryBeforeConnectionReturnsError(t *testing.T) {
	if _, err := Query[int]("RETURN 1;", nil); err == nil {
		t.Fatal("uninitialized query accepted")
	}
}

func TestUUIDsAreValidAndUnique(t *testing.T) {
	first, second := GetUUID(), GetUUID()
	if _, err := uuid.Parse(first); err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("duplicate generated IDs")
	}
}

func TestStorageHelpersBeforeConnectionReturnErrors(t *testing.T) {
	id := models.RecordID{Table: "User", ID: "missing"}
	user := &dbtype.User{Name: "User"}
	if _, err := Create(user); err == nil {
		t.Fatal("uninitialized create accepted")
	}
	if _, err := Update(id, user); err == nil {
		t.Fatal("uninitialized update accepted")
	}
	if _, err := Upsert(user); err == nil {
		t.Fatal("uninitialized upsert accepted")
	}
	if _, err := Select[dbtype.User](id.String()); err == nil {
		t.Fatal("uninitialized select accepted")
	}
	if err := Delete(id); err == nil {
		t.Fatal("uninitialized delete accepted")
	}
	if err := Relate(&id, &id, "Relation", nil); err == nil {
		t.Fatal("uninitialized relation accepted")
	}
}
