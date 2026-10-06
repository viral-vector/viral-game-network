package handler

import (
	"errors"
	"reflect"
	"testing"
)

func TestLobbyLiveWriterSkipsSnapshotOverlap(t *testing.T) {
	previous := `{"id":{"Table":"Lobby_Message","ID":"previous"},"body":"same body"}`
	next := `{"id":{"Table":"Lobby_Message","ID":"next"},"body":"same body"}`
	legacy := `{"body":"legacy message without an ID"}`
	var delivered []string
	write := lobbyLiveWriter([]string{previous, legacy}, func(raw []byte) error {
		delivered = append(delivered, string(raw))
		return nil
	})
	for _, raw := range []string{previous, next, legacy, `{"error":"invalid input"}`} {
		if err := write([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{next, legacy, `{"error":"invalid input"}`}; !reflect.DeepEqual(delivered, want) {
		t.Fatalf("live messages: %v; want %v", delivered, want)
	}
}

func TestLobbyLiveWriterReturnsDisconnectError(t *testing.T) {
	disconnected := errors.New("connection closed")
	write := lobbyLiveWriter(nil, func([]byte) error { return disconnected })
	if err := write([]byte(`{"body":"hello"}`)); !errors.Is(err, disconnected) {
		t.Fatalf("write error lost: %v", err)
	}
}
