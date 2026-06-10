package consumers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
)

type fakeNotificationEventService struct {
	statusChangedCalls    int
	deadlineReminderCalls int
	submittedCalls        int
	lastSubmitted         events.ApplicationSubmittedEvent
}

func (s *fakeNotificationEventService) HandleApplicationStatusChanged(_ context.Context, _ events.ApplicationStatusChangedEvent) error {
	s.statusChangedCalls++
	return nil
}

func (s *fakeNotificationEventService) HandleApplicationSubmitted(_ context.Context, event events.ApplicationSubmittedEvent) error {
	s.submittedCalls++
	s.lastSubmitted = event
	return nil
}

func (s *fakeNotificationEventService) HandleApplicationDeadlineReminder(_ context.Context, _ events.ApplicationDeadlineReminderEvent) error {
	s.deadlineReminderCalls++
	return nil
}

func TestNotificationEventConsumer_HandlesApplicationSubmittedEvent(t *testing.T) {
	svc := &fakeNotificationEventService{}
	consumer := NewNotificationEventConsumer(svc)

	event := events.ApplicationSubmittedEvent{
		EventID:         "evt-submit-1",
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		CitizenUserID:   "citizen-1",
		ServiceName:     "Cap CCCD",
		OccurredAt:      time.Now(),
	}
	body, _ := json.Marshal(event)

	if err := consumer.HandleApplicationSubmitted(context.Background(), body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.submittedCalls != 1 {
		t.Fatalf("expected one submitted call, got %d", svc.submittedCalls)
	}
	if svc.lastSubmitted.ApplicationCode != "HS001" {
		t.Fatalf("expected HS001, got %q", svc.lastSubmitted.ApplicationCode)
	}
}

func TestNotificationEventConsumer_HandlesDeadlineReminderEvent(t *testing.T) {
	svc := &fakeNotificationEventService{}
	consumer := NewNotificationEventConsumer(svc)

	event := events.ApplicationDeadlineReminderEvent{
		ApplicationID:   "app-1",
		ApplicationCode: "HS001",
		RecipientUserID: "staff-1",
		ServiceName:     "Cap CCCD",
		DueAt:           time.Now().Add(24 * time.Hour),
		ReminderType:    "due_in_48h",
	}
	body, _ := json.Marshal(event)

	if err := consumer.HandleApplicationDeadlineReminder(context.Background(), body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.deadlineReminderCalls != 1 {
		t.Fatalf("expected one deadline reminder call, got %d", svc.deadlineReminderCalls)
	}
}
