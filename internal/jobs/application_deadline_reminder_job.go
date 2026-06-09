package jobs

import (
	"context"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

const ReminderTypeDueIn48h = "due_in_48h"

type dueApplicationRepo interface {
	ListDueWithin(now, until time.Time) ([]models.Application, error)
}

type reminderLogRepo interface {
	Exists(applicationID, recipientUserID, reminderType string) (bool, error)
	Create(log *models.ApplicationReminderLog) error
}

type reminderPublisher interface {
	PublishApplicationDeadlineReminder(ctx context.Context, event events.ApplicationDeadlineReminderEvent) error
}

type ApplicationDeadlineReminderJob struct {
	appRepo   dueApplicationRepo
	logRepo   reminderLogRepo
	publisher reminderPublisher
	now       func() time.Time
}

func NewApplicationDeadlineReminderJob(appRepo dueApplicationRepo, logRepo reminderLogRepo, publisher reminderPublisher, now func() time.Time) *ApplicationDeadlineReminderJob {
	if now == nil {
		now = time.Now
	}
	return &ApplicationDeadlineReminderJob{appRepo: appRepo, logRepo: logRepo, publisher: publisher, now: now}
}

func (j *ApplicationDeadlineReminderJob) Run(ctx context.Context) (Result, error) {
	startedAt := j.now()
	result := Result{JobName: "application_deadline_reminder", StartedAt: startedAt}

	apps, err := j.appRepo.ListDueWithin(startedAt, startedAt.Add(48*time.Hour))
	if err != nil {
		result.Errors++
		result.FinishedAt = j.now()
		return result, err
	}

	result.Scanned = len(apps)
	for _, app := range apps {
		recipients := reminderRecipients(app)
		if len(recipients) == 0 {
			result.Skipped++
			continue
		}
		result.Matched++
		for _, recipient := range recipients {
			exists, err := j.logRepo.Exists(app.ID, recipient.userID, ReminderTypeDueIn48h)
			if err != nil {
				result.Errors++
				continue
			}
			if exists {
				result.Skipped++
				continue
			}
			if app.DueAt == nil {
				result.Skipped++
				continue
			}

			event := events.ApplicationDeadlineReminderEvent{
				ApplicationID:   app.ID,
				ApplicationCode: app.ApplicationCode,
				RecipientUserID: recipient.userID,
				RecipientRole:   recipient.role,
				ServiceName:     app.ServiceType.Name,
				DueAt:           *app.DueAt,
				ReminderType:    ReminderTypeDueIn48h,
			}
			if err := j.publisher.PublishApplicationDeadlineReminder(ctx, event); err != nil {
				result.Errors++
				continue
			}
			if err := j.logRepo.Create(&models.ApplicationReminderLog{
				ApplicationID:   app.ID,
				RecipientUserID: recipient.userID,
				ReminderType:    ReminderTypeDueIn48h,
				SentAt:          j.now(),
				CreatedAt:       j.now(),
			}); err != nil {
				result.Errors++
				continue
			}
			result.Published++
		}
	}

	result.FinishedAt = j.now()
	return result, nil
}

type reminderRecipient struct {
	userID string
	role   string
}

func reminderRecipients(app models.Application) []reminderRecipient {
	recipients := make([]reminderRecipient, 0, 2)
	if app.AssignedStaffUserID != nil && *app.AssignedStaffUserID != "" {
		recipients = append(recipients, reminderRecipient{userID: *app.AssignedStaffUserID, role: "staff"})
	}
	if app.ServiceType.ResponsibleDepartment != nil && app.ServiceType.ResponsibleDepartment.LeaderUserID != nil && *app.ServiceType.ResponsibleDepartment.LeaderUserID != "" {
		recipients = append(recipients, reminderRecipient{userID: *app.ServiceType.ResponsibleDepartment.LeaderUserID, role: "department_leader"})
	}
	return recipients
}
