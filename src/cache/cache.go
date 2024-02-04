package cache

import (
	"os"
	"log"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	log.Println("Cache Initialize!")

	rdb = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("CACHE_ENDPOINT"),
        Password: os.Getenv("CACHE_PASSWORD"),
        DB:       0,
    })
    
	log.Println("Cache Connected!")
}

func Get(key string) string {
	val, err := rdb.Get(ctx, key).Result()
    if err != nil {
        panic(err)
    }
	return val
}

func Set(key string, val string, dur time.Duration) {
	err := rdb.Set(ctx, key, val, dur).Err()
    if err != nil {
        panic(err)
    } 
}