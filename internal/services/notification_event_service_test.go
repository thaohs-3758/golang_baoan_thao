package services

import (
	"context"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

func TestNotificationEventService_HandleApplicationSubmitted_CreatesReceivedNotification(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationEventService(repo)

	event := events.ApplicationSubmittedEvent{
		EventID:         "evt-submit-1",
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		CitizenUserID:   "citizen-1",
		ServiceName:     "Cap CCCD",
		OccurredAt:      time.Now(),
	}

	if err := svc.HandleApplicationSubmitted(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected notification to be created")
	}
	if repo.created.Type != models.NotificationTypeReceived {
		t.Fatalf("expected received type, got %q", repo.created.Type)
	}
	if repo.created.SourceEventID == nil || *repo.created.SourceEventID != "evt-submit-1" {
		t.Fatal("expected source event id to be set")
	}
}

func TestNotificationEventService_HandleApplicationSubmitted_DuplicateEventIsSafe(t *testing.T) {
	repo := &fakeNotificationRepo{createErr: gorm.ErrDuplicatedKey}
	svc := NewNotificationEventService(repo)

	event := events.ApplicationSubmittedEvent{
		EventID:         "evt-submit-1",
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		CitizenUserID:   "citizen-1",
		ServiceName:     "Cap CCCD",
		OccurredAt:      time.Now(),
	}

	if err := svc.HandleApplicationSubmitted(context.Background(), event); err != nil {
		t.Fatalf("expected duplicate create to be swallowed, got %v", err)
	}
}

func TestNotificationEventService_HandleApplicationStatusChanged_CreatesNotification(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationEventService(repo)

	event := events.ApplicationStatusChangedEvent{
		EventID:         "evt-1",
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		CitizenUserID:   "citizen-1",
		ServiceName:     "Cap CCCD",
		OldStatus:       "received",
		NewStatus:       "processing",
		OccurredAt:      time.Now(),
	}

	if err := svc.HandleApplicationStatusChanged(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected notification to be created")
	}
	if repo.created.UserID != "citizen-1" {
		t.Fatalf("expected notification user citizen-1, got %q", repo.created.UserID)
	}
	if repo.created.Type != models.NotificationTypeSystem {
		t.Fatalf("expected notification type %q, got %q", models.NotificationTypeSystem, repo.created.Type)
	}
}

func TestNotificationEventService_HandleApplicationDeadlineReminder_CreatesNotification(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewNotificationEventService(repo)

	event := events.ApplicationDeadlineReminderEvent{
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		RecipientUserID: "staff-1",
		RecipientRole:   "staff",
		ServiceName:     "Cap CCCD",
		DueAt:           time.Now().Add(24 * time.Hour),
		ReminderType:    "due_in_48h",
	}

	if err := svc.HandleApplicationDeadlineReminder(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected deadline reminder notification to be created")
	}
	if repo.created.UserID != "staff-1" {
		t.Fatalf("expected notification user staff-1, got %q", repo.created.UserID)
	}
	if repo.created.Type != models.NotificationTypeDeadlineReminder {
		t.Fatalf("expected notification type %q, got %q", models.NotificationTypeDeadlineReminder, repo.created.Type)
	}
}
