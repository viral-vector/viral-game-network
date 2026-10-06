package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var renewLock = redis.NewScript(`if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('PEXPIRE', KEYS[1], ARGV[2]) else return 0 end`)
var releaseLock = redis.NewScript(`if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) else return 0 end`)

// WithLock atomically acquires a renewable lease. A lost lease cancels work's
// context; cleanup only removes the caller's own lease, never a successor's.
func WithLock(parent context.Context, key string, ttl time.Duration, work func(context.Context) error) (held bool, result error) {
	if ttl < 3*time.Millisecond || work == nil {
		return false, fmt.Errorf("invalid lock parameters")
	}
	client := rdb
	owner := uuid.NewString()
	acquired, err := client.SetNX(parent, key, owner, ttl).Result()
	if err != nil || !acquired {
		return false, err
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	stop := make(chan struct{})
	renewed := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				renewed <- nil
				return
			case <-ctx.Done():
				renewed <- ctx.Err()
				return
			case <-ticker.C:
				request, done := context.WithTimeout(ctx, ttl/3)
				updated, err := renewLock.Run(request, client, []string{key}, owner, ttl.Milliseconds()).Int()
				done()
				if err != nil || updated != 1 {
					if err == nil {
						err = fmt.Errorf("lock ownership lost: %s", key)
					}
					cancel()
					renewed <- err
					return
				}
			}
		}
	}()
	defer func() {
		close(stop)
		renewalErr := <-renewed
		request, done := context.WithTimeout(context.Background(), ttl/3)
		defer done()
		releaseErr := releaseLock.Run(request, client, []string{key}, owner).Err()
		result = errors.Join(result, renewalErr, releaseErr)
	}()
	return true, work(ctx)
}
