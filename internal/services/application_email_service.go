package services

import (
	"errors"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

var ErrApplicationEmailUnsupportedStatus = errors.New("application_email.unsupported_status")

type ApplicationEmailService struct {
	mailer Mailer
}

func NewApplicationEmailService(mailer Mailer) *ApplicationEmailService {
	return &ApplicationEmailService{mailer: mailer}
}

func (s *ApplicationEmailService) SendCitizenStatusChangeEmail(app *models.Application, newStatus models.ApplicationStatus, note string, attachmentURLs []string) error {
	if s == nil || s.mailer == nil || app == nil || strings.TrimSpace(app.CitizenUser.Email) == "" {
		return nil
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
		return ErrApplicationEmailUnsupportedStatus
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

	return s.mailer.Send(ccEmail, app.CitizenUser.Email, subject, body)
}
