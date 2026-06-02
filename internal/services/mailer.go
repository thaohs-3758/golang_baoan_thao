package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	gomail "gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(ccEmail, toEmail, subject, body string, attachments ...string) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

func LoadSMTPConfigFromEnv() SMTPConfig {
	host := utils.EnvOr("SMTP_HOST", "smtp.gmail.com")
	port, err := strconv.Atoi(utils.EnvOr("SMTP_PORT", "587"))
	if err != nil {
		log.Fatalf("invalid SMTP_PORT: %v", err)
	}
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")
	if user == "" || pass == "" {
		log.Fatal("SMTP_USERNAME and SMTP_PASSWORD must be set (use a Gmail App Password)")
	}
	from := utils.EnvOr("SMTP_FROM", user)
	fromName := utils.EnvOr("SMTP_FROM_NAME", "Cổng dịch vụ công")

	return SMTPConfig{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
		FromName: fromName,
	}
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(ccEmail, toEmail, subject, body string, attachments ...string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.cfg.From, m.cfg.FromName)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Cc", ccEmail)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)
	for _, attachment := range attachments {
		msg.Attach(attachment)
	}

	dialer := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.Username, m.cfg.Password)
	dialer.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: m.cfg.Host,
	}

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("send smtp mail: %w", err)
	}
	return nil
}
