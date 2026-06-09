package middlewares

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
)

type RateLimitStore interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error)
}

type RateLimitConfig struct {
	Name   string
	Limit  int64
	Window time.Duration
}

type RedisRateLimitStore struct {
	client *redis.Client
}

func NewRedisRateLimitStore(client *redis.Client) *RedisRateLimitStore {
	return &RedisRateLimitStore{client: client}
}

func (s *RedisRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if s == nil || s.client == nil {
		return 0, 0, errors.New("rate limit redis client is nil")
	}

	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}

	if count == 1 {
		if err := s.client.Expire(ctx, key, window).Err(); err != nil {
			return 0, 0, err
		}
		return count, window, nil
	}

	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if ttl < 0 {
		if err := s.client.Expire(ctx, key, window).Err(); err != nil {
			return 0, 0, err
		}
		ttl = window
	}

	return count, ttl, nil
}

func RateLimitMiddleware(store RateLimitStore, cfg RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if store == nil || cfg.Name == "" || cfg.Limit <= 0 || cfg.Window <= 0 {
				return next(c)
			}

			key := rateLimitKey(cfg.Name, clientIP(c))
			count, retryAfter, err := store.Increment(c.Request().Context(), key, cfg.Window)
			if err != nil {
				c.Logger().Warn("rate limit store error", "error", err)
				return next(c)
			}

			if count > cfg.Limit {
				if retryAfter <= 0 {
					retryAfter = cfg.Window
				}
				c.Response().Header().Set(echo.HeaderRetryAfter, retryAfterSeconds(retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, "common.too_many_requests")
			}

			return next(c)
		}
	}
}

func AuthLoginRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{Name: "auth_login", Limit: 5, Window: time.Minute}
}

func AuthRegisterRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{Name: "auth_register", Limit: 3, Window: time.Minute}
}

func AuthRefreshRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{Name: "auth_refresh", Limit: 20, Window: time.Minute}
}

func rateLimitKey(name, ip string) string {
	if ip == "" {
		ip = "unknown"
	}
	return "rate:" + name + ":ip:" + ip
}

func clientIP(c *echo.Context) string {
	if ip := strings.TrimSpace(c.RealIP()); ip != "" {
		return ip
	}

	remoteAddr := strings.TrimSpace(c.Request().RemoteAddr)
	if remoteAddr == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}

	return remoteAddr
}

func retryAfterSeconds(d time.Duration) string {
	seconds := int64(d.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
