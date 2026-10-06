package pubsub

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

var ctx = context.Background()
var rdb *redis.Client

func init() { Configure(os.Getenv("CACHE_ENDPOINT"), os.Getenv("CACHE_PASSWORD")) }

// Configure installs a Redis connection; call before serving requests.
func Configure(endpoint, password string) {
	if rdb != nil {
		rdb.Close()
	}
	rdb = redis.NewClient(&redis.Options{Addr: endpoint, Password: password, DB: 1})
}

func Pub(key string, msg []byte) error {
	err := rdb.Publish(ctx, key, msg).Err()
	if err != nil {
		return err
	}
	return nil
}

func Sub(key string) *redis.PubSub {
	sb := rdb.Subscribe(ctx, key)
	return sb
}

func Close(sb *redis.PubSub) {
	sb.Close()
}
