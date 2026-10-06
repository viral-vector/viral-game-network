package cache_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/tests/support"
)

func TestLockHasOneConcurrentOwner(t *testing.T) {
	support.Redis(t)
	var wg sync.WaitGroup
	var owners atomic.Int32
	start := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{}, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := cache.WithLock(context.Background(), "lock", time.Second, func(context.Context) error { owners.Add(1); <-release; return nil })
			if err != nil {
				t.Error(err)
			}
			finished <- struct{}{}
		}()
	}
	close(start)
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	timedOut := false
	for i := 0; i < 31; i++ {
		select {
		case <-finished:
		case <-deadline.C:
			timedOut = true
		}
		if timedOut {
			break
		}
	}
	close(release)
	wg.Wait()
	if timedOut || owners.Load() != 1 {
		t.Fatalf("lock admitted %d concurrent owners", owners.Load())
	}
}

func TestLostLockCancelsWorkWithoutDeletingSuccessor(t *testing.T) {
	server := support.Redis(t)
	held, err := cache.WithLock(context.Background(), "lock", 300*time.Millisecond, func(ctx context.Context) error {
		server.Set("lock", "successor")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			return fmt.Errorf("lost lease did not cancel work")
		}
	})
	if !held || err == nil {
		t.Fatalf("lease loss unreported: held=%v error=%v", held, err)
	}
	if value, err := server.Get("lock"); err != nil || value != "successor" {
		t.Fatal("expired owner deleted successor", value, err)
	}
}

func TestLockRenewalAndErrorCleanup(t *testing.T) {
	server := support.Redis(t)
	held, err := cache.WithLock(context.Background(), "lock", 300*time.Millisecond, func(context.Context) error {
		server.FastForward(250 * time.Millisecond)
		time.Sleep(150 * time.Millisecond)
		server.FastForward(100 * time.Millisecond)
		if !server.Exists("lock") {
			t.Error("active lease was not renewed")
		}
		return fmt.Errorf("work failed")
	})
	if !held || err == nil || server.Exists("lock") {
		t.Fatalf("failed work did not clean up: held=%v error=%v", held, err)
	}
}
