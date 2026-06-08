package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/queue"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if err := configs.LoadI18nMessages("locales"); err != nil {
		log.Fatalf("failed to load i18n: %v", err)
	}

	db := configs.InitDB()

	rabbitConn, err := queue.NewRabbitMQConnection()
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer rabbitConn.Close()

	applicationRepo := repositories.NewApplicationRepository(db)
	smtpCfg := services.LoadSMTPConfigFromEnv()
	mailer := services.NewSMTPMailer(smtpCfg)
	emailSvc := services.NewApplicationEmailService(mailer)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	consumer := queue.NewRabbitMQConsumer(rabbitConn)

	if err := consumer.ConsumeApplicationStatusChanged(
		ctx,
		"application.status_changed.email",
		func(ctx context.Context, body []byte) error {
			var event events.ApplicationStatusChangedEvent
			if err := json.Unmarshal(body, &event); err != nil {
				return err
			}

			app, err := applicationRepo.GetByID(event.ApplicationID)
			if err != nil {
				return err
			}

			if app == nil {
				return nil
			}

			attachmentURLs := make([]string, 0, len(app.ApplicationAttachments))
			for _, attachment := range app.ApplicationAttachments {
				if strings.TrimSpace(attachment.FileURL) == "" {
					continue
				}
				attachmentURLs = append(attachmentURLs, attachment.FileURL)
			}

			return emailSvc.SendCitizenStatusChangeEmail(
				app,
				models.ApplicationStatus(event.NewStatus),
				event.Note,
				attachmentURLs,
			)
		},
	); err != nil {
		log.Fatalf("failed to start email consumer: %v", err)
	}

	log.Println("worker started")
	<-ctx.Done()
	log.Println("worker stopped")
}
