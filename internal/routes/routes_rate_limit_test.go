package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/labstack/echo/v5"
)

type routeRateLimitStore struct {
	counts map[string]int64
}

func newRouteRateLimitStore() *routeRateLimitStore {
	return &routeRateLimitStore{counts: map[string]int64{}}
}

func (s *routeRateLimitStore) Increment(_ context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	s.counts[key]++
	return s.counts[key], window, nil
}

func newRoutesEchoForRateLimit(store middlewares.RateLimitStore) *echo.Echo {
	e := echo.New()
	e.Renderer = &fakeRenderer{}
	SetupRoutes(e, &ApiHandler{
		AdminAuthHandler:        &handlers.AdminAuthHandler{},
		AdminDashboardHandler:   &handlers.AdminDashboardHandler{},
		AdminUserHandler:        &handlers.AdminUserHandler{},
		AdminDepartmentHandler:  &handlers.AdminDepartmentHandler{},
		AdminCategoryHandler:    &handlers.AdminCategoryHandler{},
		AdminApplicationHandler: &handlers.AdminApplicationHandler{},
		AdminLogHandler:         &handlers.AdminLogHandler{},
		AdminCitizenHandler:     &handlers.AdminCitizenHandler{},
		AdminProfileHandler:     &handlers.AdminProfileHandler{},
		AuthHandler:             &handlers.AuthHandler{},
		CitizenProfileHandler:   &handlers.CitizenProfileHandler{},
		ServiceCatalogHandler:   &handlers.ServiceCatalogHandler{},
		ApplicationHandler:      &handlers.ApplicationHandler{},
		NotificationHandler:     &handlers.NotificationHandler{},
		CitizenWebHandler:       &handlers.CitizenWebHandler{},
		RealtimeHandler:         &handlers.RealtimeHandler{},
		RateLimitStore:          store,
	})
	return e
}

func TestPublicAuthRoutesAreRateLimited(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		limit       int64
		limiterName string
	}{
		{name: "citizen web login", method: http.MethodPost, path: "/login", limit: 5, limiterName: "auth_login"},
		{name: "citizen web register", method: http.MethodPost, path: "/register", limit: 3, limiterName: "auth_register"},
		{name: "admin web login", method: http.MethodPost, path: "/admin/login", limit: 5, limiterName: "auth_login"},
		{name: "api login", method: http.MethodPost, path: "/api/auth/login", limit: 5, limiterName: "auth_login"},
		{name: "api register", method: http.MethodPost, path: "/api/auth/register", limit: 3, limiterName: "auth_register"},
		{name: "api refresh", method: http.MethodPost, path: "/api/auth/refresh", limit: 20, limiterName: "auth_refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newRouteRateLimitStore()
			e := newRoutesEchoForRateLimit(store)
			store.counts["rate:"+tt.limiterName+":ip:203.0.113.55"] = tt.limit

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = "203.0.113.55:1234"
			last := httptest.NewRecorder()
			e.ServeHTTP(last, req)
			if last.Code != http.StatusTooManyRequests {
				t.Fatalf("expected %s %s to be rate limited with %d, got %d", tt.method, tt.path, http.StatusTooManyRequests, last.Code)
			}
			if last.Header().Get(echo.HeaderRetryAfter) == "" {
				t.Fatal("expected Retry-After header")
			}
		})
	}
}

func TestRateLimitStoreIsOptionalForExistingRouteTests(t *testing.T) {
	e := newRoutesEchoForRateLimit(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusTooManyRequests {
		t.Fatal("expected missing store not to rate-limit routes")
	}
}
