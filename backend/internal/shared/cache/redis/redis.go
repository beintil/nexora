package cache_redis

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/shared/cache"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
}

func NewCacheRedis(client *redis.Client) cache.Cache {
	return &Redis{client: client}
}

func (r Redis) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("cache_redis/Set: %w", err)
	}
	return nil
}

func (r Redis) Get(ctx context.Context, key string, desc any) error {
	err := r.client.Get(ctx, key).Scan(desc)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return cache.ErrorCacheValueNotFound
		}
		return fmt.Errorf("cache_redis/Get: %w", err)
	}
	return nil
}

func (r Redis) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("cache_redis/Delete: %w", err)
	}
	return nil
}
