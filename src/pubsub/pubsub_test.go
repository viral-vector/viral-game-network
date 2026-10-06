package pubsub_test

import (
	"context"
	"testing"
	"time"
	"viral-game-network/src/pubsub"
	"viral-game-network/tests/support"
)

func TestPublishOnlyToSubscribedChannel(t *testing.T) {
	support.Redis(t)
	sub := pubsub.Sub("lobby:test")
	defer pubsub.Close(sub)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pubsub.Pub("other", []byte("wrong")); err != nil {
		t.Fatal(err)
	}
	if err := pubsub.Pub("lobby:test", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil || msg.Payload != "hello" || msg.Channel != "lobby:test" {
		t.Fatalf("%+v %v", msg, err)
	}
}
