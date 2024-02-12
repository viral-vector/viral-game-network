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

func Get(key string) (interface{}, error) {
	dtype, err_o := rdb.Type(ctx, key).Result()
	if err_o != nil {
		return nil, err_o
	}

	var (
		val interface{}
		err error
	)
	switch dtype {
		case "list":
			length := int64(0)
			length, err = rdb.LLen(ctx, key).Result()
			if err != nil {
				return nil, err
			}
			val, err = rdb.LRange(ctx, key,  0, length-1).Result()
			if err != nil {
				return nil, err
			}
		default:
			val, err = rdb.Get(ctx, key).Result()
			if err != nil {
				return nil, err
			}
	}
	
	return val, nil
}

func Set(key string, val string, dur time.Duration) error {
	err := rdb.Set(ctx, key, val, dur).Err()
    if err != nil {
		return err
    }
	return nil 
}

func Add(key string, val string) error {
	err := rdb.RPush(ctx, key, val).Err()
    if err != nil {
		return err
    }
	return nil 
}

func Exp(guid string, dur time.Duration) error{
	err := rdb.Expire(ctx, guid, dur).Err()
	if err != nil {
		return err
	}
	return nil
}

func Del(guid string) error{
	err := rdb.Del(ctx, guid).Err()
	if err != nil {
		return err
	}
	return nil
}