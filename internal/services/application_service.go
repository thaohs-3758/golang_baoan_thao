package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"gorm.io/gorm"
)

const (
	MaxTotalAttachmentBytes = 30 * 1024 * 1024 // 30 MiB — also used as server body limit
	maxAttachments          = 10
	maxFileSizeBytes        = 10 * 1024 * 1024 // 10 MiB per file
)

var (
	ErrApplicationInvalidInput = errors.New("application.invalid_input")
	ErrApplicationNotFound     = errors.New("application.not_found")
	ErrServiceTypeNotFound     = errors.New("application.service_type_not_found")
	ErrServiceTypeInactive     = errors.New("application.service_type_inactive")
	ErrMissingRequiredField    = errors.New("application.missing_required_field")
	ErrInvalidSubmittedData    = errors.New("application.invalid_submitted_data")
	ErrInvalidFormSchema       = errors.New("application.invalid_form_schema")
	ErrAttachmentRequired      = errors.New("application.attachment_required")
	ErrSupplementNotAllowed    = errors.New("application.supplement_not_allowed")
	ErrTooManyAttachments      = errors.New("application.attachment_too_many")
	ErrAttachmentTooLarge      = errors.New("application.attachment_too_large")
	ErrAttachmentInvalidType   = errors.New("application.attachment_invalid_type")
)

type ApplicationService struct {
	appRepo            repositories.ApplicationRepository
	serviceTypeRepo    repositories.ServiceTypeRepository
	userRepo           repositories.UserRepository
	citizenProfileRepo repositories.CitizenProfileRepository
	storage            utils.FileStorage
	mailer             Mailer
	activityLogger     activityLogger
}

type activityLogger interface {
	Log(log *models.ActivityLog) error
}

func NewApplicationService(
	appRepo repositories.ApplicationRepository,
	serviceTypeRepo repositories.ServiceTypeRepository,
	userRepo repositories.UserRepository,
	citizenProfileRepo repositories.CitizenProfileRepository,
	storage utils.FileStorage,
	mailer Mailer,
	loggers ...activityLogger,
) *ApplicationService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &ApplicationService{
		appRepo:            appRepo,
		serviceTypeRepo:    serviceTypeRepo,
		userRepo:           userRepo,
		citizenProfileRepo: citizenProfileRepo,
		storage:            storage,
		mailer:             mailer,
		activityLogger:     logger,
	}
}

func (s *ApplicationService) SubmitApplication(
	citizenUserID string,
	req *dtos.SubmitApplicationRequest,
	files []*multipart.FileHeader,
) (*dtos.ApplicationResponse, error) {
	if req == nil {
		return nil, ErrApplicationInvalidInput
	}
	st, err := s.serviceTypeRepo.GetByID(context.Background(), req.ServiceTypeID)
	if err != nil {
		return nil, ErrServiceTypeNotFound
	}
	if st == nil {
		return nil, ErrServiceTypeNotFound
	}
	if !st.IsActive {
		return nil, ErrServiceTypeInactive
	}

	if err := validateSubmittedData(req.SubmittedData, st.FormSchema); err != nil {
		return nil, err
	}

	if err := validateAttachmentLimits(files); err != nil {
		return nil, err
	}

	now := time.Now()
	app := &models.Application{
		ApplicationCode: utils.GenerateApplicationCode(),
		CitizenUserID:   citizenUserID,
		ServiceTypeID:   req.ServiceTypeID,
		Status:          models.ApplicationStatusReceived,
		SubmittedData:   req.SubmittedData,
		SubmittedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tmpID := utils.GenerateUUID()
	app.ID = tmpID

	atts := make([]models.ApplicationAttachment, 0, len(files))

	savedURLs := make([]string, 0, len(files))
	for _, fh := range files {
		pubURL, mimeType, size, err := s.storage.SaveApplicationFile(tmpID, fh)
		if err != nil {
			_ = s.storage.RemoveApplicationDir(tmpID)
			if errors.Is(err, utils.ErrDisallowedMime) {
				return nil, ErrAttachmentInvalidType
			}
			return nil, fmt.Errorf("save attachment: %w", err)
		}
		savedURLs = append(savedURLs, pubURL)
		sz := size
		atts = append(atts, models.ApplicationAttachment{
			UploadedByUserID: citizenUserID,
			FileName:         fh.Filename,
			FileURL:          pubURL,
			FileType:         mimeType,
			FileSize:         &sz,
			AttachmentType:   models.AttachmentTypeSubmitted,
			CreatedAt:        now,
		})
	}

	loc := configs.DefaultLocale
	notifParams := map[string]string{"code": app.ApplicationCode, "service": st.Name}
	notif := &models.Notification{
		UserID:    citizenUserID,
		Title:     configs.TLang(loc, "notification.received.title", notifParams),
		Message:   configs.TLang(loc, "notification.received.message", notifParams),
		Type:      models.NotificationTypeReceived,
		CreatedAt: now,
	}

	if err := s.appRepo.CreateWithAttachments(app, atts, notif, utils.GenerateApplicationCode); err != nil {
		_ = s.storage.RemoveApplicationDir(tmpID)
		return nil, fmt.Errorf("create application: %w", err)
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &citizenUserID,
		Action:      "application.submit",
		EntityType:  "application",
		EntityID:    &app.ID,
		Description: "Nộp hồ sơ: " + app.ApplicationCode,
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"service_type_id": app.ServiceTypeID,
			"status":          app.Status,
			"attachments":     len(atts),
		}),
		CreatedAt: now,
	})

	go func() {
		user, err := s.userRepo.FindByID(citizenUserID)
		if err != nil || user == nil {
			return
		}
		if s.citizenProfileRepo != nil {
			profile, err := s.citizenProfileRepo.GetByUserID(citizenUserID)
			if err == nil && profile != nil && !profile.EmailNotificationEnabled {
				return
			}
		}
		citizenName := user.Name
		if citizenName == "" {
			citizenName = configs.TLang("vi", "common.default_citizen_name", nil)
		}
		params := map[string]string{
			"name":         citizenName,
			"service":      st.Name,
			"code":         app.ApplicationCode,
			"submitted_at": app.SubmittedAt.Format("15:04 02/01/2006"),
			"from_name":    utils.EnvOr("SMTP_FROM_NAME", configs.TLang("vi", "common.app_name", nil)),
		}
		subject := configs.TLang("vi", "mail.application_received.subject", params)
		body := configs.TLang("vi", "mail.application_received.body", params)
		if err := s.mailer.Send(user.Email, subject, body); err != nil {
			log.Printf("send confirmation email failed: %v", err)
		}
	}()

	return toApplicationResponse(app, st, atts), nil
}

func (s *ApplicationService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func (s *ApplicationService) ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.ListByCitizen(userID, page, limit)
}

func (s *ApplicationService) GetMyApplication(userID, appID string) (*dtos.ApplicationResponse, error) {
	app, err := s.appRepo.GetByIDForCitizen(appID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("get application: %w", err)
	}
	return toApplicationResponseFromModel(app), nil
}

func (s *ApplicationService) ListMyApplicationStatusHistory(userID, appID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	app, err := s.appRepo.GetByIDForCitizen(appID, userID)
	if err != nil {
		if isApplicationRecordNotFound(err) {
			return nil, 0, ErrApplicationNotFound
		}
		return nil, 0, fmt.Errorf("get application: %w", err)
	}
	if app == nil {
		return nil, 0, ErrApplicationNotFound
	}
	return s.appRepo.ListStatusLogsByCitizen(appID, userID, page, limit, since)
}

func (s *ApplicationService) UploadMyApplicationSupplements(userID, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
	if len(files) == 0 {
		return nil, ErrAttachmentRequired
	}

	if err := validateAttachmentLimits(files); err != nil {
		return nil, err
	}

	app, err := s.appRepo.GetByIDForCitizen(appID, userID)
	if err != nil {
		if isApplicationRecordNotFound(err) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("get application: %w", err)
	}
	if app == nil {
		return nil, ErrApplicationNotFound
	}

	if app.Status != models.ApplicationStatusProcessing && app.Status != models.ApplicationStatusNeedMoreInfo {
		return nil, ErrSupplementNotAllowed
	}

	now := time.Now()
	atts := make([]models.ApplicationAttachment, 0, len(files))
	savedURLs := make([]string, 0, len(files))
	for _, fh := range files {
		pubURL, mimeType, size, saveErr := s.storage.SaveApplicationFile(app.ID, fh)
		if saveErr != nil {
			for _, u := range savedURLs {
				_ = s.storage.RemoveFile(u)
			}
			if errors.Is(saveErr, utils.ErrDisallowedMime) {
				return nil, ErrAttachmentInvalidType
			}
			return nil, fmt.Errorf("save supplement attachment: %w", saveErr)
		}
		savedURLs = append(savedURLs, pubURL)
		sz := size
		atts = append(atts, models.ApplicationAttachment{
			UploadedByUserID: userID,
			FileName:         fh.Filename,
			FileURL:          pubURL,
			FileType:         mimeType,
			FileSize:         &sz,
			AttachmentType:   models.AttachmentTypeSupplement,
			CreatedAt:        now,
		})
	}

	if err := s.appRepo.CreateAttachments(app.ID, atts); err != nil {
		for _, a := range atts {
			_ = s.storage.RemoveFile(a.FileURL)
		}
		return nil, fmt.Errorf("create supplement attachments: %w", err)
	}

	resp := make([]dtos.ApplicationAttachmentResponse, 0, len(atts))
	for _, a := range atts {
		resp = append(resp, dtos.ApplicationAttachmentResponse{
			ID:             a.ID,
			FileName:       a.FileName,
			FileURL:        a.FileURL,
			FileType:       a.FileType,
			FileSize:       a.FileSize,
			AttachmentType: string(a.AttachmentType),
			CreatedAt:      a.CreatedAt,
		})
	}

	return resp, nil
}

func validateSubmittedData(data json.RawMessage, schema json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	required := extractRequiredFields(schema)
	if len(required) == 0 {
		return nil
	}

	var submitted map[string]interface{}
	if err := json.Unmarshal(data, &submitted); err != nil {
		return ErrInvalidSubmittedData
	}
	for _, field := range required {
		v, ok := submitted[field]
		if !ok || v == nil {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
		}
		if str, ok := v.(string); ok && strings.TrimSpace(str) == "" {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
		}
	}
	return nil
}

// extractRequiredFields handles both the legacy flat format and the new i18n field format.
func extractRequiredFields(schema json.RawMessage) []string {
	// New format: {"fields":[{"key":"x","required":true,...}]}
	var newFmt struct {
		Fields []struct {
			Key      string `json:"key"`
			Required bool   `json:"required"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(schema, &newFmt); err == nil && len(newFmt.Fields) > 0 && newFmt.Fields[0].Key != "" {
		var out []string
		for _, f := range newFmt.Fields {
			if f.Required {
				out = append(out, f.Key)
			}
		}
		return out
	}
	// Legacy format: {"required":["x"],"required_fields":["x"]}
	var legacy struct {
		Required       []string `json:"required"`
		RequiredFields []string `json:"required_fields"`
	}
	if err := json.Unmarshal(schema, &legacy); err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range append(legacy.Required, legacy.RequiredFields...) {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return out
}

func isApplicationRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "record not found")
}

func validateAttachmentLimits(files []*multipart.FileHeader) error {
	if len(files) > maxAttachments {
		return ErrTooManyAttachments
	}
	var total int64
	for _, f := range files {
		if f.Size > maxFileSizeBytes {
			return ErrAttachmentTooLarge
		}
		total += f.Size
	}
	if total > MaxTotalAttachmentBytes {
		return ErrAttachmentTooLarge
	}
	return nil
}

func toApplicationResponse(app *models.Application, st *models.ServiceType, atts []models.ApplicationAttachment) *dtos.ApplicationResponse {
	attResponses := make([]dtos.ApplicationAttachmentResponse, 0, len(atts))
	for _, a := range atts {
		attResponses = append(attResponses, dtos.ApplicationAttachmentResponse{
			ID:             a.ID,
			FileName:       a.FileName,
			FileURL:        a.FileURL,
			FileType:       a.FileType,
			FileSize:       a.FileSize,
			AttachmentType: string(a.AttachmentType),
			CreatedAt:      a.CreatedAt,
		})
	}
	return &dtos.ApplicationResponse{
		ID:              app.ID,
		ApplicationCode: app.ApplicationCode,
		ServiceTypeID:   app.ServiceTypeID,
		ServiceTypeName: st.Name,
		Status:          string(app.Status),
		SubmittedData:   app.SubmittedData,
		SubmittedAt:     app.SubmittedAt,
		Attachments:     attResponses,
	}
}

func toApplicationResponseFromModel(app *models.Application) *dtos.ApplicationResponse {
	attResponses := make([]dtos.ApplicationAttachmentResponse, 0)
	for _, a := range app.ApplicationAttachments {
		attResponses = append(attResponses, dtos.ApplicationAttachmentResponse{
			ID:             a.ID,
			FileName:       a.FileName,
			FileURL:        a.FileURL,
			FileType:       a.FileType,
			FileSize:       a.FileSize,
			AttachmentType: string(a.AttachmentType),
			CreatedAt:      a.CreatedAt,
		})
	}
	return &dtos.ApplicationResponse{
		ID:              app.ID,
		ApplicationCode: app.ApplicationCode,
		ServiceTypeID:   app.ServiceTypeID,
		ServiceTypeName: app.ServiceType.Name,
		Status:          string(app.Status),
		RejectedReason:  app.RejectedReason,
		SubmittedData:   app.SubmittedData,
		SubmittedAt:     app.SubmittedAt,
		Attachments:     attResponses,
	}
}
