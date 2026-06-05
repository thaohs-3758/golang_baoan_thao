package cache

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestJSONCache_AllowsNilContext(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:0",
		DialTimeout: time.Millisecond,
	})
	c := NewJSONCache(rdb, time.Minute)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("cache methods should not panic with nil context: %v", r)
		}
	}()

	var dest map[string]string
	_ = c.Get(nil, "cache:test", &dest)
	c.Set(nil, "cache:test", map[string]string{"ok": "true"})
	c.DeletePattern(nil, "cache:*")
}
