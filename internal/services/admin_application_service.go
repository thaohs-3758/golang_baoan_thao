package services

import (
	"errors"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
)

var ErrAdminApplicationNotFound = errors.New("application.not_found")
var ErrAdminApplicationInvalidTransition = errors.New("application.invalid_transition")
var ErrAdminApplicationRejectReasonRequired = errors.New("application.reject_reason_required")
var ErrAdminApplicationNeedMoreInfoNoteRequired = errors.New("application.need_more_info_note_required")

type AdminApplicationService struct {
	appRepo          repositories.ApplicationRepository
	assignService    *ApplicationAssignmentService
	storage          utils.FileStorage
	activityLogger   activityLogger
	notificationRepo repositories.NotificationRepository
	mailer           Mailer
}

func NewAdminApplicationService(appRepo repositories.ApplicationRepository, assignService *ApplicationAssignmentService, storage utils.FileStorage, loggers ...activityLogger) *AdminApplicationService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &AdminApplicationService{appRepo: appRepo, assignService: assignService, storage: storage, activityLogger: logger}
}

func (s *AdminApplicationService) WithNotificationRepo(repo repositories.NotificationRepository) *AdminApplicationService {
	s.notificationRepo = repo
	return s
}

func (s *AdminApplicationService) WithMailer(mailer Mailer) *AdminApplicationService {
	s.mailer = mailer
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

	if err := s.appRepo.ProcessStatusUpdate(app.ID, &app.Status, newStatus, resultNote, rejectedReason, processingStartedAt, completedAt, processedBy, attachments); err != nil {
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

	s.notifyCitizenStatusChange(app, newStatus, note, now)
	s.sendCitizenStatusChangeEmail(app, newStatus, note, savedURLs)

	return nil
}

func (s *AdminApplicationService) sendCitizenStatusChangeEmail(app *models.Application, newStatus models.ApplicationStatus, note string, attachmentURLs []string) {
	if s.mailer == nil || app == nil || strings.TrimSpace(app.CitizenUser.Email) == "" {
		return
	}

	params := map[string]string{
		"code":    app.ApplicationCode,
		"service": app.ServiceType.Name,
		"note":    strings.TrimSpace(note),
	}

	var subjectKey, bodyKey string
	switch newStatus {
	case models.ApplicationStatusProcessing:
		subjectKey = "notification.processing.title"
		bodyKey = "notification.processing.message"
	case models.ApplicationStatusNeedMoreInfo:
		subjectKey = "notification.need_more_info.title"
		bodyKey = "notification.need_more_info.message"
	case models.ApplicationStatusApproved:
		subjectKey = "notification.approved.title"
		bodyKey = "notification.approved.message"
	case models.ApplicationStatusRejected:
		subjectKey = "notification.rejected.title"
		bodyKey = "notification.rejected.message"
	default:
		return
	}

	ccEmail := ""
	if app.ServiceType.ResponsibleDepartment != nil &&
		app.ServiceType.ResponsibleDepartment.LeaderUser != nil &&
		strings.TrimSpace(app.ServiceType.ResponsibleDepartment.LeaderUser.Email) != "" {
		ccEmail = app.ServiceType.ResponsibleDepartment.LeaderUser.Email
	}

	subject := configs.TLang(configs.DefaultLocale, subjectKey, params)
	body := configs.TLang(configs.DefaultLocale, bodyKey, params)
	if len(attachmentURLs) > 0 {
		body += "\n\nAttachment URLs:"
		for _, attachmentURL := range attachmentURLs {
			if strings.TrimSpace(attachmentURL) == "" {
				continue
			}
			body += "\n- " + attachmentURL
		}
	}
	if err := s.mailer.Send(ccEmail, app.CitizenUser.Email, subject, body); err != nil {
		log.Printf("send status-change email failed for app %s: %v", app.ID, err)
	}
}

func (s *AdminApplicationService) notifyCitizenStatusChange(app *models.Application, newStatus models.ApplicationStatus, note string, now time.Time) {
	if s.notificationRepo == nil || app == nil {
		return
	}
	loc := configs.DefaultLocale
	params := map[string]string{
		"code":    app.ApplicationCode,
		"service": app.ServiceType.Name,
		"note":    strings.TrimSpace(note),
	}

	var titleKey, messageKey string
	var notifType models.NotificationType
	switch newStatus {
	case models.ApplicationStatusProcessing:
		titleKey = "notification.processing.title"
		messageKey = "notification.processing.message"
		notifType = models.NotificationTypeSystem
	case models.ApplicationStatusNeedMoreInfo:
		titleKey = "notification.need_more_info.title"
		messageKey = "notification.need_more_info.message"
		notifType = models.NotificationTypeNeedMoreInfo
	case models.ApplicationStatusApproved:
		titleKey = "notification.approved.title"
		messageKey = "notification.approved.message"
		notifType = models.NotificationTypeResult
	case models.ApplicationStatusRejected:
		titleKey = "notification.rejected.title"
		messageKey = "notification.rejected.message"
		notifType = models.NotificationTypeResult
	default:
		return
	}

	appID := app.ID
	notif := &models.Notification{
		UserID:        app.CitizenUserID,
		ApplicationID: &appID,
		Title:         configs.TLang(loc, titleKey, params),
		Message:       configs.TLang(loc, messageKey, params),
		Type:          notifType,
		CreatedAt:     now,
	}
	if err := s.notificationRepo.Create(notif); err != nil {
		log.Printf("create status-change notification failed for app %s: %v", app.ID, err)
	}
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
