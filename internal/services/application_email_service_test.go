package services

import (
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/stretchr/testify/assert"
)

type statusChangeMailer struct {
	ccEmail string
	toEmail string
	subject string
	body    string
	err     error
	sent    bool
}

func (m *statusChangeMailer) Send(ccEmail, toEmail, subject, body string, _ ...string) error {
	m.sent = true
	m.ccEmail = ccEmail
	m.toEmail = toEmail
	m.subject = subject
	m.body = body
	return m.err
}

var _ Mailer = (*statusChangeMailer)(nil)

func TestApplicationEmailService_SendCitizenStatusChangeEmail_SendsWithLeaderCCAndAttachments(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	leader := &models.User{Email: "leader@test.com"}
	app := &models.Application{
		ID:              "app-1",
		ApplicationCode: "APP-1",
		CitizenUser:     models.User{Email: "citizen@test.com"},
		ServiceType: models.ServiceType{
			Name: "Test Service",
			ResponsibleDepartment: &models.Department{
				LeaderUser: leader,
			},
		},
	}
	mailer := &statusChangeMailer{}
	svc := NewApplicationEmailService(mailer)

	err := svc.SendCitizenStatusChangeEmail(app, models.ApplicationStatusApproved, "approved", []string{"", "/uploads/result.pdf"})

	assert.NoError(t, err)
	assert.True(t, mailer.sent)
	assert.Equal(t, "leader@test.com", mailer.ccEmail)
	assert.Equal(t, "citizen@test.com", mailer.toEmail)
	assert.NotEmpty(t, mailer.subject)
	assert.Contains(t, mailer.body, "/uploads/result.pdf")
	assert.NotContains(t, mailer.body, "\n- \n")
}

func TestApplicationEmailService_SendCitizenStatusChangeEmail_ReturnsMailerError(t *testing.T) {
	app := &models.Application{
		ApplicationCode: "APP-1",
		CitizenUser:     models.User{Email: "citizen@test.com"},
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}
	mailerErr := errors.New("smtp down")
	svc := NewApplicationEmailService(&statusChangeMailer{err: mailerErr})

	err := svc.SendCitizenStatusChangeEmail(app, models.ApplicationStatusApproved, "", nil)

	assert.ErrorIs(t, err, mailerErr)
}

func TestApplicationEmailService_SendCitizenStatusChangeEmail_SkipsMissingMailerOrEmail(t *testing.T) {
	app := &models.Application{CitizenUser: models.User{Email: " "}}
	mailer := &statusChangeMailer{}

	assert.NoError(t, NewApplicationEmailService(nil).SendCitizenStatusChangeEmail(app, models.ApplicationStatusApproved, "", nil))
	assert.NoError(t, NewApplicationEmailService(mailer).SendCitizenStatusChangeEmail(app, models.ApplicationStatusApproved, "", nil))
	assert.False(t, mailer.sent)
}
