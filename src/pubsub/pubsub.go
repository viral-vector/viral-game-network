package pubsub

import (
	"os"
	"log"
	//"time"
	"context"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	log.Println("PubSub Initialize!")

	rdb = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("CACHE_ENDPOINT"),
        Password: os.Getenv("CACHE_PASSWORD"),
        DB:       1,
    })
    
	log.Println("PubSub Connected!")
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
