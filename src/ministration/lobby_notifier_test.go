package ministration

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/pubsub"
	"viral-game-network/tests/support"
)

func TestNotifyPublishesAndCachesMessage(t *testing.T) {
	server := support.Redis(t)
	channel := "lobby:Lobby:game:channel"
	sub := pubsub.Sub(channel)
	defer pubsub.Close(sub)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Service_Lobby_Notify("Lobby:game", "hello", "User:alice"); err != nil {
		t.Fatal(err)
	}
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var decoded dbtype.LobbyMessage
	if err := json.Unmarshal([]byte(msg.Payload), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.User_ID != "User:alice" || decoded.Body != "hello" || decoded.Date_Created == "" {
		t.Fatalf("%+v", decoded)
	}
	history, err := server.List(channel)
	if err != nil || len(history) != 1 || history[0] != msg.Payload || server.TTL(channel) != time.Hour {
		t.Fatalf("history %v, TTL %v, error %v", history, server.TTL(channel), err)
	}
}

func TestNotifyReturnsCacheFailure(t *testing.T) {
	server := support.Redis(t)
	channel := "lobby:Lobby:game:channel"
	sub := pubsub.Sub(channel)
	defer pubsub.Close(sub)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	server.Set("lobby:Lobby:game:channel", "wrong type")
	if err := Service_Lobby_Notify("Lobby:game", "hello", ""); err == nil {
		t.Fatal("cache error was swallowed")
	}
	if message, err := sub.ReceiveTimeout(ctx, 50*time.Millisecond); err == nil {
		t.Fatalf("failed notification was broadcast: %+v", message)
	}
}

func TestNotifyBoundsReplayHistory(t *testing.T) {
	server := support.Redis(t)
	for i := 0; i < 150; i++ {
		if err := Service_Lobby_Notify("Lobby:game", "hello", "User:alice"); err != nil {
			t.Fatal(err)
		}
	}
	history, err := server.List("lobby:Lobby:game:channel")
	if err != nil || len(history) != 100 {
		t.Fatalf("replay history not bounded: length=%d error=%v", len(history), err)
	}
	ids := make(map[string]bool)
	for _, raw := range history {
		var message dbtype.LobbyMessage
		if err := json.Unmarshal([]byte(raw), &message); err != nil {
			t.Fatal(err)
		}
		id := message.ModelID()
		if id == "" || ids[id] {
			t.Fatal("replay messages lack unique identities")
		}
		ids[id] = true
	}
}
