package cache

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LobbyHistoryLimit bounds both replay traffic and per-lobby Redis storage.
const LobbyHistoryLimit = 100

var storeAndPublish = redis.NewScript(`
redis.call('RPUSH', KEYS[1], ARGV[1])
redis.call('LTRIM', KEYS[1], -tonumber(ARGV[2]), -1)
redis.call('PEXPIRE', KEYS[1], ARGV[3])
return redis.call('PUBLISH', KEYS[1], ARGV[1])
`)

// PublishHistory stores a message before publishing it, in one Redis operation.
// A wrong-type history key fails before any subscriber sees the message.
func PublishHistory(channel, message string, ttl time.Duration) error {
	if ttl < time.Millisecond {
		return fmt.Errorf("invalid history lifetime")
	}
	return storeAndPublish.Run(ctx, rdb, []string{channel}, message, LobbyHistoryLimit, ttl.Milliseconds()).Err()
}
