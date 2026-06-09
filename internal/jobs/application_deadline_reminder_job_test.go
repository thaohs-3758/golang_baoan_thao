package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

func TestApplicationDeadlineReminderJobPublishesForStaffAndManager(t *testing.T) {
	now := time.Now()
	staffID := "staff-1"
	managerID := "manager-1"
	apps := []models.Application{
		{
			ID:                  "app-1",
			ApplicationCode:     "HS001",
			Status:              models.ApplicationStatusProcessing,
			DueAt:               ptrTimeJob(now.Add(24 * time.Hour)),
			AssignedStaffUserID: &staffID,
			ServiceType: models.ServiceType{
				Name: "Cap CCCD",
				ResponsibleDepartment: &models.Department{
					LeaderUserID: &managerID,
				},
			},
		},
	}

	pub := &fakeReminderPublisher{}
	job := NewApplicationDeadlineReminderJob(
		&fakeDueApplicationRepo{apps: apps},
		&fakeReminderLogRepo{},
		pub,
		func() time.Time { return now },
	)

	result, err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Published != 2 {
		t.Fatalf("expected 2 reminders, got %d", result.Published)
	}
	if len(pub.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(pub.events))
	}
}

type fakeDueApplicationRepo struct {
	apps []models.Application
	err  error
}

func (r *fakeDueApplicationRepo) ListDueWithin(_, _ time.Time) ([]models.Application, error) {
	return r.apps, r.err
}

type fakeReminderLogRepo struct {
	exists map[string]bool
	logs   []*models.ApplicationReminderLog
	err    error
}

func (r *fakeReminderLogRepo) Exists(applicationID, recipientUserID, reminderType string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	if r.exists == nil {
		return false, nil
	}
	return r.exists[applicationID+":"+recipientUserID+":"+reminderType], nil
}

func (r *fakeReminderLogRepo) Create(log *models.ApplicationReminderLog) error {
	if r.err != nil {
		return r.err
	}
	r.logs = append(r.logs, log)
	return nil
}

type fakeReminderPublisher struct {
	events []events.ApplicationDeadlineReminderEvent
	err    error
}

func (p *fakeReminderPublisher) PublishApplicationDeadlineReminder(_ context.Context, event events.ApplicationDeadlineReminderEvent) error {
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, event)
	return nil
}

func ptrTimeJob(v time.Time) *time.Time { return &v }
