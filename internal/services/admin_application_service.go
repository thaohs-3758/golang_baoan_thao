package services

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/realtime"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/google/uuid"
)

var ErrAdminApplicationNotFound = errors.New("application.not_found")
var ErrAdminApplicationInvalidTransition = errors.New("application.invalid_transition")
var ErrAdminApplicationRejectReasonRequired = errors.New("application.reject_reason_required")
var ErrAdminApplicationNeedMoreInfoNoteRequired = errors.New("application.need_more_info_note_required")

type RealtimeNotifier interface {
	SendToUser(userID string, msg realtime.Message)
}

type ApplicationEventPublisher interface {
	PublishApplicationSubmitted(ctx context.Context, event events.ApplicationSubmittedEvent) error
	PublishApplicationStatusChanged(ctx context.Context, event events.ApplicationStatusChangedEvent) error
}

type AdminApplicationService struct {
	appRepo          repositories.ApplicationRepository
	assignService    *ApplicationAssignmentService
	storage          utils.FileStorage
	activityLogger   activityLogger
	mailer           Mailer
	emailService     *ApplicationEmailService
	realTimeNotifier RealtimeNotifier
	eventPublisher   ApplicationEventPublisher
}

func NewAdminApplicationService(appRepo repositories.ApplicationRepository, assignService *ApplicationAssignmentService, storage utils.FileStorage, loggers ...activityLogger) *AdminApplicationService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &AdminApplicationService{appRepo: appRepo, assignService: assignService, storage: storage, activityLogger: logger}
}

func (s *AdminApplicationService) WithEventPublisher(publisher ApplicationEventPublisher) *AdminApplicationService {
	s.eventPublisher = publisher
	return s
}

func (s *AdminApplicationService) WithMailer(mailer Mailer) *AdminApplicationService {
	s.mailer = mailer
	s.emailService = NewApplicationEmailService(mailer)
	return s
}

func (s *AdminApplicationService) WithRealtimeNotifier(notifier RealtimeNotifier) *AdminApplicationService {
	s.realTimeNotifier = notifier
	return s
}

func (s *AdminApplicationService) ListApplications(filter repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.AdminList(filter, page, limit)
}

func (s *AdminApplicationService) ListApplicationsForActor(filter repositories.ApplicationFilter, page, limit int, role models.UserRole, actorID string) ([]models.Application, int64, error) {
	if role == models.UserRoleStaff {
		filter.AssignedStaffUserID = actorID
	}
	return s.appRepo.AdminList(filter, page, limit)
}

func (s *AdminApplicationService) GetApplication(id string) (*models.Application, error) {
	return s.appRepo.GetByID(id)
}

func (s *AdminApplicationService) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	if err := s.assignService.AssignApplicationToStaff(applicationID, toStaffUserID, assignedBy); err != nil {
		return err
	}
	return nil
}

func (s *AdminApplicationService) ProcessApplication(applicationID string, newStatus models.ApplicationStatus, note string, files []*multipart.FileHeader, processedBy string) error {
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil || app == nil {
		return ErrAdminApplicationNotFound
	}
	oldStatus := app.Status

	if !isAllowedAdminTransition(app.Status, newStatus) {
		return ErrAdminApplicationInvalidTransition
	}
	if newStatus == models.ApplicationStatusRejected && strings.TrimSpace(note) == "" {
		return ErrAdminApplicationRejectReasonRequired
	}
	if newStatus == models.ApplicationStatusNeedMoreInfo && strings.TrimSpace(note) == "" {
		return ErrAdminApplicationNeedMoreInfoNoteRequired
	}

	now := time.Now()
	processingStartedAt := app.ProcessingStartedAt
	completedAt := app.CompletedAt
	if newStatus == models.ApplicationStatusProcessing && processingStartedAt == nil {
		processingStartedAt = &now
	}
	if newStatus == models.ApplicationStatusApproved || newStatus == models.ApplicationStatusRejected {
		completedAt = &now
	}

	resultNote := note
	rejectedReason := ""
	if newStatus == models.ApplicationStatusRejected {
		rejectedReason = note
		resultNote = ""
	}

	attachments := make([]models.ApplicationAttachment, 0, len(files))
	savedURLs := make([]string, 0, len(files))
	if len(files) > 0 {
		if s.storage == nil {
			return errors.New("storage not configured")
		}
		for _, fh := range files {
			pubURL, mimeType, size, saveErr := s.storage.SaveApplicationFile(app.ID, fh)
			if saveErr != nil {
				for _, u := range savedURLs {
					_ = s.storage.RemoveFile(u)
				}
				return saveErr
			}
			savedURLs = append(savedURLs, pubURL)
			sz := size
			attachments = append(attachments, models.ApplicationAttachment{
				UploadedByUserID: processedBy,
				FileName:         fh.Filename,
				FileURL:          pubURL,
				FileType:         mimeType,
				FileSize:         &sz,
				AttachmentType:   models.AttachmentTypeResult,
				CreatedAt:        now,
			})
		}
	}

	if err := s.appRepo.ProcessStatusUpdate(app.ID, &oldStatus, newStatus, resultNote, rejectedReason, processingStartedAt, completedAt, processedBy, attachments); err != nil {
		if s.storage != nil {
			for _, a := range attachments {
				_ = s.storage.RemoveFile(a.FileURL)
			}
		}
		return err
	}
	actorID := processedBy
	s.logActivity(&models.ActivityLog{
		ActorUserID: &actorID,
		Action:      "application.status_update",
		EntityType:  "application",
		EntityID:    &app.ID,
		Description: "Cập nhật trạng thái hồ sơ " + app.ApplicationCode + " → " + string(newStatus),
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"changes": map[string]any{
				"status": map[string]any{
					"before": app.Status,
					"after":  newStatus,
				},
			},
			"result_note_set":     strings.TrimSpace(resultNote) != "",
			"rejected_reason_set": strings.TrimSpace(rejectedReason) != "",
			"attachments":         len(attachments),
		}),
		CreatedAt: now,
	})

	s.notifyCitizenStatusChangeRealtime(app, newStatus, note)

	event := events.ApplicationStatusChangedEvent{
		EventID:         uuid.NewString(),
		ApplicationID:   app.ID,
		ApplicationCode: app.ApplicationCode,
		CitizenUserID:   app.CitizenUserID,
		ServiceName:     app.ServiceType.Name,
		OldStatus:       string(oldStatus),
		NewStatus:       string(newStatus),
		Note:            note,
		ChangedBy:       processedBy,
		OccurredAt:      time.Now(),
	}

	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishApplicationStatusChanged(context.Background(), event); err != nil {
			log.Printf("failed to publish application status changed event: %v", err)
		}
	}

	return nil
}

func (s *AdminApplicationService) notifyCitizenStatusChangeRealtime(app *models.Application, newStatus models.ApplicationStatus, note string) {
	if s.realTimeNotifier == nil || app == nil {
		return
	}

	params := map[string]string{
		"code":    app.ApplicationCode,
		"service": app.ServiceType.Name,
		"note":    strings.TrimSpace(note),
	}

	var messageKey string

	switch newStatus {
	case models.ApplicationStatusProcessing:
		messageKey = "notification.processing.message"
	case models.ApplicationStatusNeedMoreInfo:
		messageKey = "notification.need_more_info.message"
	case models.ApplicationStatusApproved:
		messageKey = "notification.approved.message"
	case models.ApplicationStatusRejected:
		messageKey = "notification.rejected.message"
	default:
		return
	}

	s.realTimeNotifier.SendToUser(app.CitizenUserID, realtime.Message{
		Type:            "application_status_changed",
		UserID:          app.CitizenUserID,
		ApplicationID:   app.ID,
		ApplicationCode: app.ApplicationCode,
		Status:          string(newStatus),
		Message:         configs.TLang(configs.DefaultLocale, messageKey, params),
	})
}

func (s *AdminApplicationService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}

func isAllowedAdminTransition(current, next models.ApplicationStatus) bool {
	switch current {
	case models.ApplicationStatusReceived:
		return next == models.ApplicationStatusProcessing || next == models.ApplicationStatusNeedMoreInfo
	case models.ApplicationStatusProcessing:
		return next == models.ApplicationStatusNeedMoreInfo || next == models.ApplicationStatusApproved || next == models.ApplicationStatusRejected
	case models.ApplicationStatusNeedMoreInfo:
		return next == models.ApplicationStatusProcessing || next == models.ApplicationStatusApproved || next == models.ApplicationStatusRejected
	default:
		return false
	}
}
