package consumers

import (
	"context"
	"encoding/json"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
)

type notificationEventService interface {
	HandleApplicationSubmitted(ctx context.Context, event events.ApplicationSubmittedEvent) error
	HandleApplicationStatusChanged(ctx context.Context, event events.ApplicationStatusChangedEvent) error
	HandleApplicationDeadlineReminder(ctx context.Context, event events.ApplicationDeadlineReminderEvent) error
}

type NotificationEventConsumer struct {
	svc notificationEventService
}

func NewNotificationEventConsumer(svc notificationEventService) *NotificationEventConsumer {
	return &NotificationEventConsumer{svc: svc}
}

func (c *NotificationEventConsumer) HandleApplicationSubmitted(ctx context.Context, body []byte) error {
	var event events.ApplicationSubmittedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	return c.svc.HandleApplicationSubmitted(ctx, event)
}

func (c *NotificationEventConsumer) HandleApplicationStatusChanged(ctx context.Context, body []byte) error {
	var event events.ApplicationStatusChangedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	return c.svc.HandleApplicationStatusChanged(ctx, event)
}

func (c *NotificationEventConsumer) HandleApplicationDeadlineReminder(ctx context.Context, body []byte) error {
	var event events.ApplicationDeadlineReminderEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	return c.svc.HandleApplicationDeadlineReminder(ctx, event)
}
