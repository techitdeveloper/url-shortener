package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheService(client *redis.Client, ttlSeconds int) *CacheService {
	return &CacheService{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *CacheService) Set(ctx context.Context, key, value string) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *CacheService) BuildURLKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}
