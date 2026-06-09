package services

import (
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type fakeNotificationRepo struct {
	listItems    []models.Notification
	listTotal    int64
	listErr      error
	markReadErr  error
	markAllErr   error
	createErr    error
	created      *models.Notification
	countUnread  int64
	countErr     error
}

func (r *fakeNotificationRepo) ListByUserID(_ string, _ repositories.NotificationFilter, _, _ int) ([]models.Notification, int64, error) {
	return r.listItems, r.listTotal, r.listErr
}

func (r *fakeNotificationRepo) MarkAsRead(_, _ string) error {
	return r.markReadErr
}

func (r *fakeNotificationRepo) MarkAllAsRead(_ string) error {
	return r.markAllErr
}

func (r *fakeNotificationRepo) Create(notif *models.Notification) error {
	r.created = notif
	return r.createErr
}

func (r *fakeNotificationRepo) CountUnread(_ string) (int64, error) {
	return r.countUnread, r.countErr
}

var _ repositories.NotificationRepository = (*fakeNotificationRepo)(nil)

func TestNotificationService_List_Success(t *testing.T) {
	now := time.Now()
	repo := &fakeNotificationRepo{
		listItems: []models.Notification{
			{ID: "n1", UserID: "u1", Title: "Hello", Message: "World", Type: models.NotificationTypeSystem, CreatedAt: now},
		},
		listTotal: 1,
	}
	svc := NewNotificationService(repo)

	items, total, err := svc.List("u1", repositories.NotificationFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "n1", items[0].ID)
	assert.Equal(t, "Hello", items[0].Title)
}

func TestNotificationService_List_Error(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeNotificationRepo{listErr: repoErr}
	svc := NewNotificationService(repo)

	items, total, err := svc.List("u1", repositories.NotificationFilter{}, 1, 10)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

func TestNotificationService_List_Empty(t *testing.T) {
	repo := &fakeNotificationRepo{listItems: []models.Notification{}, listTotal: 0}
	svc := NewNotificationService(repo)

	items, total, err := svc.List("u1", repositories.NotificationFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, items, 0)
}

func TestNotificationService_MarkAsRead_Success(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationService(repo)

	err := svc.MarkAsRead("n1", "u1")
	assert.NoError(t, err)
}

func TestNotificationService_MarkAsRead_Error(t *testing.T) {
	repoErr := errors.New("mark error")
	repo := &fakeNotificationRepo{markReadErr: repoErr}
	svc := NewNotificationService(repo)

	err := svc.MarkAsRead("n1", "u1")
	assert.ErrorIs(t, err, repoErr)
}

func TestNotificationService_MarkAllAsRead_Success(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationService(repo)

	err := svc.MarkAllAsRead("u1")
	assert.NoError(t, err)
}

func TestNotificationService_MarkAllAsRead_Error(t *testing.T) {
	repoErr := errors.New("mark all error")
	repo := &fakeNotificationRepo{markAllErr: repoErr}
	svc := NewNotificationService(repo)

	err := svc.MarkAllAsRead("u1")
	assert.ErrorIs(t, err, repoErr)
}

func TestNotificationService_CountUnread_Success(t *testing.T) {
	repo := &fakeNotificationRepo{countUnread: 5}
	svc := NewNotificationService(repo)

	count, err := svc.CountUnread("u1")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestNotificationService_CountUnread_Error(t *testing.T) {
	repoErr := errors.New("count error")
	repo := &fakeNotificationRepo{countErr: repoErr}
	svc := NewNotificationService(repo)

	count, err := svc.CountUnread("u1")
	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, int64(0), count)
}

func TestNotificationServiceCreate(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationService(repo)

	notif := &models.Notification{
		UserID:    "user-1",
		Title:     "Deadline soon",
		Message:   "Application HS001 is due within 48 hours.",
		Type:      models.NotificationTypeDeadlineReminder,
		CreatedAt: time.Now(),
	}

	if err := svc.Create(notif); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected notification to be persisted")
	}
}
