// Package redis provides simple Redis cache functions.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kirsle/go-website-template/webapp/log"
)

var ctx = context.Background()

var Client *redis.Client

/*
Setup the Redis connection.

The addr format is like:

- localhost:6379
- localhost:6379/6

The latter format to specify the DB number if not the default (0).
*/
func Setup(addr string) error {
	// Parse the addr string.
	parts := strings.Split(addr, "/")
	addr = parts[0]
	db := 0
	if len(parts) > 1 && len(parts[1]) > 0 {
		a, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("redis DB number was not an integer: %s", err)
		}
		db = a
	}

	Client = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	return nil
}

// Set a JSON serializable object in Redis.
func Set(key string, v interface{}, expire time.Duration) error {
	bin, err := json.Marshal(v)
	if err != nil {
		return err
	}

	log.Debug("redis.Set(%s): %s", key, bin)

	_, err = Client.Set(ctx, key, bin, expire).Result()
	if err != nil {
		return err
	}

	return nil
}

// Get a JSON serialized value out of Redis.
func Get(key string, v any) error {
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	log.Debug("redis.Get(%s): %s", key, val)
	return json.Unmarshal([]byte(val), v)
}

// Exists checks if a Redis key existed.
func Exists(key string) bool {
	val, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	log.Debug("redis.Exists(%s): %d", key, val)
	return val == 1
}

// Delete a key from Redis.
func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}
