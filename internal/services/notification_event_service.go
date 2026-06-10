package services

import (
	"context"
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"gorm.io/gorm"
)

type NotificationEventService struct {
	repo repositories.NotificationRepository
}

func NewNotificationEventService(repo repositories.NotificationRepository) *NotificationEventService {
	return &NotificationEventService{repo: repo}
}

func (s *NotificationEventService) HandleApplicationSubmitted(ctx context.Context, event events.ApplicationSubmittedEvent) error {
	_ = ctx

	params := map[string]string{
		"code":    event.ApplicationCode,
		"service": event.ServiceName,
	}
	eventID := event.EventID
	notif := &models.Notification{
		UserID:        event.CitizenUserID,
		ApplicationID: &event.ApplicationID,
		SourceEventID: &eventID,
		Title:         configs.TLang(configs.DefaultLocale, "notification.received.title", params),
		Message:       configs.TLang(configs.DefaultLocale, "notification.received.message", params),
		Type:          models.NotificationTypeReceived,
		CreatedAt:     time.Now(),
	}
	return s.persistNotification(notif)
}

func (s *NotificationEventService) HandleApplicationStatusChanged(ctx context.Context, event events.ApplicationStatusChangedEvent) error {
	_ = ctx

	titleKey, messageKey, notifType := statusChangeTemplate(event.NewStatus)
	if titleKey == "" || messageKey == "" {
		return nil
	}

	params := map[string]string{
		"code":    event.ApplicationCode,
		"service": event.ServiceName,
		"note":    event.Note,
	}
	eventID := event.EventID
	notif := &models.Notification{
		UserID:        event.CitizenUserID,
		ApplicationID: &event.ApplicationID,
		SourceEventID: &eventID,
		Title:         configs.TLang(configs.DefaultLocale, titleKey, params),
		Message:       configs.TLang(configs.DefaultLocale, messageKey, params),
		Type:          notifType,
		CreatedAt:     time.Now(),
	}
	return s.persistNotification(notif)
}

func (s *NotificationEventService) HandleApplicationDeadlineReminder(ctx context.Context, event events.ApplicationDeadlineReminderEvent) error {
	_ = ctx

	params := map[string]string{
		"code":    event.ApplicationCode,
		"service": event.ServiceName,
	}
	notif := &models.Notification{
		UserID:        event.RecipientUserID,
		ApplicationID: &event.ApplicationID,
		Title:         configs.TLang(configs.DefaultLocale, "notification.deadline_reminder.title", params),
		Message:       configs.TLang(configs.DefaultLocale, "notification.deadline_reminder.message", params),
		Type:          models.NotificationTypeDeadlineReminder,
		CreatedAt:     time.Now(),
	}
	return s.persistNotification(notif)
}

func (s *NotificationEventService) persistNotification(notif *models.Notification) error {
	err := s.repo.Create(notif)
	if errors.Is(err, gorm.ErrDuplicatedKey) && notif.SourceEventID != nil && *notif.SourceEventID != "" {
		return nil
	}
	return err
}

func statusChangeTemplate(newStatus string) (titleKey string, messageKey string, notifType models.NotificationType) {
	switch newStatus {
	case string(models.ApplicationStatusProcessing):
		return "notification.processing.title", "notification.processing.message", models.NotificationTypeSystem
	case string(models.ApplicationStatusNeedMoreInfo):
		return "notification.need_more_info.title", "notification.need_more_info.message", models.NotificationTypeNeedMoreInfo
	case string(models.ApplicationStatusApproved):
		return "notification.approved.title", "notification.approved.message", models.NotificationTypeResult
	case string(models.ApplicationStatusRejected):
		return "notification.rejected.title", "notification.rejected.message", models.NotificationTypeResult
	default:
		return "", "", ""
	}
}
