//go:build integration

package repository

import (
	"testing"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestClusterEventUpsertPreservesUnchangedDateAndUpdatesChanges(t *testing.T) {
	support.Storage(t)
	input := &dbtype.SystemEvent{Ref_ID: "event", Ref_Source: "cluster", Ref_Target: "pod", Severity: "info", Message: "Pending"}
	first, err := PopSystemevent(input)
	if err != nil {
		t.Fatal(err)
	}
	oldDate := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if _, err := database.Query[any](`UPDATE type::record($id) SET date_created = $date;`, map[string]interface{}{"id": first.ModelID(), "date": oldDate}); err != nil {
		t.Fatal(err)
	}
	unchanged, err := PopSystemevent(input)
	if err != nil || unchanged.ModelID() != first.ModelID() || unchanged.Date_Created != oldDate {
		t.Fatalf("repeat polling reset event history: %+v %v", unchanged, err)
	}
	input.Message, input.Severity = "Image pull failed", "warning"
	updated, err := PopSystemevent(input)
	if err != nil || updated.ModelID() != first.ModelID() || updated.Date_Created == oldDate || updated.Message != input.Message || updated.Severity != input.Severity {
		t.Fatalf("changed event not updated: %+v %v", updated, err)
	}
	if events, total, err := AllSystemEvents(10, 1); err != nil || len(events) != 1 || total != 1 {
		t.Fatalf("duplicate event history: %+v %d %v", events, total, err)
	}
	if events, err := SelSystemEventsInFrame(60); err != nil || len(events) != 1 {
		t.Fatalf("event update missing from live feed: %+v %v", events, err)
	}
}
