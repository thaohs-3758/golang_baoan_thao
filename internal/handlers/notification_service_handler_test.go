package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/labstack/echo/v5"
)

type fakeNotificationAPIService struct {
	items       []dtos.NotificationResponse
	total       int64
	listErr     error
	countUnread int64
	countErr    error
	markReadErr error
	markAllErr  error
}

func (s *fakeNotificationAPIService) List(_ string, _ repositories.NotificationFilter, _, _ int) ([]dtos.NotificationResponse, int64, error) {
	return s.items, s.total, s.listErr
}

func (s *fakeNotificationAPIService) CountUnread(_ string) (int64, error) {
	return s.countUnread, s.countErr
}

func (s *fakeNotificationAPIService) MarkAsRead(_, _ string) error {
	return s.markReadErr
}

func (s *fakeNotificationAPIService) MarkAllAsRead(_ string) error {
	return s.markAllErr
}

func TestNotificationServiceHandler_List(t *testing.T) {
	e := echo.New()
	svc := &fakeNotificationAPIService{
		items: []dtos.NotificationResponse{{ID: "n1", Title: "Hello"}},
		total: 1,
	}
	h := NewNotificationServiceHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/notifications?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
