package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type JSONCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewJSONCache(rdb *redis.Client, ttl time.Duration) *JSONCache {
	return &JSONCache{rdb: rdb, ttl: ttl}
}

func (c *JSONCache) Get(ctx context.Context, key string, dest any) bool {
	if c == nil || c.rdb == nil {
		return false
	}
	ctx = cacheContext(ctx)

	raw, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false
	}

	return true
}

func (c *JSONCache) Set(ctx context.Context, key string, value any) {
	if c == nil || c.rdb == nil {
		return
	}
	ctx = cacheContext(ctx)

	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	_ = c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func (c *JSONCache) DeletePattern(ctx context.Context, pattern string) {
	if c == nil || c.rdb == nil {
		return
	}
	ctx = cacheContext(ctx)

	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		_ = c.rdb.Del(ctx, iter.Val()).Err()
	}
}

func cacheContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
