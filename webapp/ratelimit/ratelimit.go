package ratelimit

import (
	"fmt"
	"time"

	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/redis"
	"github.com/kirsle/go-website-template/webapp/utility"
)

// Limiter implements a Redis-backed rate limit for logins or otherwise.
type Limiter struct {
	Namespace  string        // kind of rate limiter ("login")
	ID         interface{}   // unique ID of the resource being pinged (str or ints)
	Limit      int           // how many pings within the window period
	Window     time.Duration // the window period/expiration of Redis key
	CooldownAt int           // how many pings before the cooldown is enforced
	Cooldown   time.Duration // time to wait between fails
}

// Redis object behind the rate limiter.
type Data struct {
	Pings     int
	NotBefore time.Time
}

// Ping the rate limiter.
func (l *Limiter) Ping() error {
	var (
		key = l.Key()
		now = time.Now()
	)

	// Get stored data from Redis if any.
	var data Data
	redis.Get(key, &data)

	// Are we cooling down?
	if now.Before(data.NotBefore) {
		return fmt.Errorf(
			"You are doing that too often. Please wait %s before trying again.",
			utility.FormatDurationCoarse(data.NotBefore.Sub(now)),
		)
	}

	// Increment the ping count.
	data.Pings++

	// Have we hit the wall?
	if data.Pings >= l.Limit {
		return fmt.Errorf(
			"You have hit the rate limit; please wait the full %s before trying again.",
			utility.FormatDurationCoarse(l.Window),
		)
	}

	// Are we throttled?
	if l.CooldownAt > 0 && data.Pings > l.CooldownAt {
		data.NotBefore = now.Add(l.Cooldown)
		if err := redis.Set(key, data, l.Window); err != nil {
			return fmt.Errorf("Couldn't set Redis key for rate limiter: %s", err)
		}
		return fmt.Errorf(
			"Please wait %s before trying again. You have %d more attempt(s) remaining before you will be locked "+
				"out for %s.",
			utility.FormatDurationCoarse(l.Cooldown),
			l.Limit-data.Pings,
			utility.FormatDurationCoarse(l.Window),
		)
	}

	// Save their ping count to Redis.
	if err := redis.Set(key, data, l.Window); err != nil {
		return fmt.Errorf("Couldn't set Redis key for rate limiter: %s", err)
	}

	return nil
}

// Clear the rate limiter, cleaning up the Redis key (e.g., after successful login).
func (l *Limiter) Clear() error {
	return redis.Delete(l.Key())
}

// Key formats the Redis key.
func (l *Limiter) Key() string {
	var str string
	switch t := l.ID.(type) {
	case int:
		str = fmt.Sprintf("%d", t)
	case uint64:
		str = fmt.Sprintf("%d", t)
	case int64:
		str = fmt.Sprintf("%d", t)
	case uint32:
		str = fmt.Sprintf("%d", t)
	case int32:
		str = fmt.Sprintf("%d", t)
	default:
		str = fmt.Sprintf("%s", t)
	}
	return fmt.Sprintf(config.RateLimitRedisKey, l.Namespace, str)
}
