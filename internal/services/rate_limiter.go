package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: client,
	}
}

func (r *RateLimiter) CheckRateLimit(ctx context.Context, userID int, maxRequests int, window time.Duration) (bool, int, time.Time, error) {
	key := fmt.Sprintf("rate_limit: user: %d", userID)

	countStr, err := r.client.Get(ctx, key).Result()

	var count int
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return false, 0, time.Time{}, err
	} else {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			return false, 0, time.Time{}, err
		}
	}

	if count >= maxRequests {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			return false, count, time.Time{}, err
		}
		resetTime := time.Now().Add(ttl)
		return false, count, resetTime, nil
	}

	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)

	if count == 0 {
		pipe.Expire(ctx, key, window)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, count, time.Time{}, err
	}

	remaining := maxRequests - count - 1
	resetTime := time.Now().Add(window)

	return true, remaining, resetTime, nil
}

func (r *RateLimiter) CheckRateLimitByIP(ctx context.Context, ip string, maxRequests int, window time.Duration) (bool, int, time.Time, error) {
	key := fmt.Sprintf("rate_limit: ip: %s", ip)

	countStr, err := r.client.Get(ctx, key).Result()

	var count int
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return false, 0, time.Time{}, err
	} else {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			return false, 0, time.Time{}, err
		}
	}

	if count >= maxRequests {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			return false, 0, time.Time{}, err
		}
		resetTime := time.Now().Add(ttl)
		return false, count, resetTime, nil
	}

	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)

	if count == 0 {
		pipe.Expire(ctx, key, window)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, count, time.Time{}, err
	}

	remaining := maxRequests - count - 1
	resetTime := time.Now().Add(window)

	return true, remaining, resetTime, nil
}
