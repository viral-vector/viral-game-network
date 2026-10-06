package database

import (
	"github.com/google/uuid"
	"testing"
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
