package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("CACHE_ENDPOINT"),
		Password: os.Getenv("CACHE_PASSWORD"),
		DB:       0,
	})
}

// Get retrieves a value from Redis for the given key and attempts to convert it into type T.
// If T is a string, it returns the raw string. Otherwise, it tries to unmarshal the string as JSON.
func Get[T any](key string) (T, error) {
	var result T

	// For simplicity, we'll assume the value is stored as a string.
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return result, err
	}

	// Check if T is a string type.
	// One way to do this is to create a dummy variable of type T and use a type switch.
	var dummy T
	switch any(dummy).(type) {
	case string:
		// Directly return the string value.
		return any(val).(T), nil
	default:
		// Assume the string is JSON and unmarshal it.
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return result, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}
		return result, nil
	}
}

// Set stores a string value in Redis with an expiration duration.
func Set(key string, val string, dur time.Duration) error {
	return rdb.Set(ctx, key, val, dur).Err()
}

// Add pushes a string value into a Redis list.
func Add(key string, val string) error {
	return rdb.RPush(ctx, key, val).Err()
}

// Exp sets the expiration for a given key.
func Exp(key string, dur time.Duration) error {
	return rdb.Expire(ctx, key, dur).Err()
}

// Del deletes the specified key from Redis.
func Del(key string) error {
	return rdb.Del(ctx, key).Err()
}