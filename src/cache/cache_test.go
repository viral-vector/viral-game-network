package cache_test

import (
	"testing"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/tests/support"
)

func TestCacheRoundTripAndExpiry(t *testing.T) {
	server := support.Redis(t)
	type value struct {
		Name  string
		Count int
	}
	if err := cache.Set("profile", value{"Alice", 3}, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := cache.Get[value]("profile")
	if err != nil || got != (value{"Alice", 3}) {
		t.Fatalf("%+v %v", got, err)
	}
	if err := cache.Exp("profile", time.Second); err != nil {
		t.Fatal(err)
	}
	server.FastForward(2 * time.Second)
	if _, err := cache.Get[value]("profile"); err == nil {
		t.Fatal("expired key survived")
	}
	cache.Set("text", "raw", 0)
	if got, _ := cache.Get[string]("text"); got != "raw" {
		t.Fatal(got)
	}
	cache.Del("text")
	if server.Exists("text") {
		t.Fatal("deleted key survived")
	}
}

func TestCacheErrorsAndListOrder(t *testing.T) {
	server := support.Redis(t)
	server.Set("bad", "invalid JSON")
	if _, err := cache.Get[map[string]string]("bad"); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if err := cache.Set("channel", make(chan int), 0); err == nil {
		t.Fatal("unencodable value accepted")
	}
	cache.Add("history", "first")
	cache.Add("history", "second")
	got, err := cache.List("history")
	if err != nil || len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("%v %v", got, err)
	}
	if err := cache.Add("history", make(chan int)); err == nil {
		t.Fatal("unencodable list value accepted")
	}
}
