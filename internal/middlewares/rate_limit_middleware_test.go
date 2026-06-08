package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

type fakeRateLimitStore struct {
	counts      map[string]int64
	windows     map[string]time.Duration
	err         error
	seenKeys    []string
	seenWindows []time.Duration
}

func newFakeRateLimitStore() *fakeRateLimitStore {
	return &fakeRateLimitStore{
		counts:  map[string]int64{},
		windows: map[string]time.Duration{},
	}
}

func (s *fakeRateLimitStore) Increment(_ context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if s.err != nil {
		return 0, 0, s.err
	}
	s.seenKeys = append(s.seenKeys, key)
	s.seenWindows = append(s.seenWindows, window)
	s.counts[key]++
	s.windows[key] = window

	return s.counts[key], window, nil
}

func TestRateLimitMiddlewareAllowsRequestsWithinLimit(t *testing.T) {
	store := newFakeRateLimitStore()
	cfg := RateLimitConfig{Name: "auth_login", Limit: 2, Window: time.Minute}
	handler := RateLimitMiddleware(store, cfg)(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		c, rec := newMiddlewareContext(http.MethodPost, "/api/auth/login")
		c.Request().RemoteAddr = "192.0.2.10:" + strconv.Itoa(5000+i)

		if err := handler(c); err != nil {
			t.Fatalf("request %d expected nil error, got %v", i+1, err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}
}

func TestRateLimitMiddlewareRejectsRequestsOverLimit(t *testing.T) {
	store := newFakeRateLimitStore()
	cfg := RateLimitConfig{Name: "auth_login", Limit: 1, Window: time.Minute}
	handler := RateLimitMiddleware(store, cfg)(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		c, _ := newMiddlewareContext(http.MethodPost, "/api/auth/login")
		c.Request().RemoteAddr = "192.0.2.11:5000"

		err := handler(c)
		if i == 0 && err != nil {
			t.Fatalf("first request expected nil error, got %v", err)
		}
		if i == 1 {
			assertHTTPErrorCode(t, err, http.StatusTooManyRequests)
			httpErr := err.(*echo.HTTPError)
			if httpErr.Message != "common.too_many_requests" {
				t.Fatalf("expected common.too_many_requests message, got %v", httpErr.Message)
			}
		}
	}
}

func TestRateLimitMiddlewareUsesNameIPAndWindow(t *testing.T) {
	store := newFakeRateLimitStore()
	cfg := RateLimitConfig{Name: "auth_register", Limit: 3, Window: 30 * time.Second}
	handler := RateLimitMiddleware(store, cfg)(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	c, rec := newMiddlewareContext(http.MethodPost, "/register")
	c.Request().RemoteAddr = "203.0.113.20:4444"

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if len(store.seenKeys) != 1 {
		t.Fatalf("expected one key, got %d", len(store.seenKeys))
	}
	if store.seenKeys[0] != "rate:auth_register:ip:203.0.113.20" {
		t.Fatalf("unexpected key %q", store.seenKeys[0])
	}
	if store.seenWindows[0] != 30*time.Second {
		t.Fatalf("expected 30s window, got %s", store.seenWindows[0])
	}
}

func TestRateLimitMiddlewareSeparatesLimiterNames(t *testing.T) {
	store := newFakeRateLimitStore()
	login := RateLimitMiddleware(store, RateLimitConfig{Name: "auth_login", Limit: 10, Window: time.Minute})(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	register := RateLimitMiddleware(store, RateLimitConfig{Name: "auth_register", Limit: 10, Window: time.Minute})(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	c1, _ := newMiddlewareContext(http.MethodPost, "/login")
	c1.Request().RemoteAddr = "198.51.100.7:1000"
	if err := login(c1); err != nil {
		t.Fatalf("login expected nil error, got %v", err)
	}

	c2, _ := newMiddlewareContext(http.MethodPost, "/register")
	c2.Request().RemoteAddr = "198.51.100.7:1000"
	if err := register(c2); err != nil {
		t.Fatalf("register expected nil error, got %v", err)
	}

	if store.counts["rate:auth_login:ip:198.51.100.7"] != 1 {
		t.Fatal("expected login key count 1")
	}
	if store.counts["rate:auth_register:ip:198.51.100.7"] != 1 {
		t.Fatal("expected register key count 1")
	}
}

func TestRateLimitMiddlewareFailsOpenWhenStoreErrors(t *testing.T) {
	store := newFakeRateLimitStore()
	store.err = errors.New("redis down")
	cfg := RateLimitConfig{Name: "auth_login", Limit: 1, Window: time.Minute}

	called := false
	handler := RateLimitMiddleware(store, cfg)(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusAccepted)
	})

	c, rec := newMiddlewareContext(http.MethodPost, "/api/auth/login")
	c.Request().RemoteAddr = "192.0.2.12:5000"

	if err := handler(c); err != nil {
		t.Fatalf("expected fail-open nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}
