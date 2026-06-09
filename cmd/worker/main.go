package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	"github.com/awesome-academy/golang_baoan_thao/internal/jobs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/queue"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/scheduler"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
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
	reminderLogRepo := repositories.NewApplicationReminderLogRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationSvc := services.NewNotificationService(notificationRepo)
	uploadDir := utils.EnvOr("UPLOAD_DIR", "./uploads")
	storage := utils.NewLocalDiskStorage(uploadDir, "/uploads")
	smtpCfg := services.LoadSMTPConfigFromEnv()
	mailer := services.NewSMTPMailer(smtpCfg)
	emailSvc := services.NewApplicationEmailService(mailer)
	publisher := queue.NewRabbitMQPublisher(rabbitConn)

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

	if err := consumer.ConsumeApplicationDeadlineReminder(
		ctx,
		"application.deadline_reminder.notification",
		func(ctx context.Context, body []byte) error {
			var event events.ApplicationDeadlineReminderEvent
			if err := json.Unmarshal(body, &event); err != nil {
				return err
			}

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
			return notificationSvc.Create(notif)
		},
	); err != nil {
		log.Fatalf("failed to start deadline reminder consumer: %v", err)
	}

	s := scheduler.New()
	if err := s.Register("application_deadline_reminder", "0 * * * *", jobs.NewApplicationDeadlineReminderJob(applicationRepo, reminderLogRepo, publisher, time.Now).Run); err != nil {
		log.Fatalf("failed to register reminder job: %v", err)
	}
	if err := s.Register("temporary_file_cleanup", "0 * * * *", jobs.NewTemporaryFileCleanupJob(storage, applicationRepo, 24*time.Hour, time.Now).Run); err != nil {
		log.Fatalf("failed to register cleanup job: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		log.Fatalf("failed to start scheduler: %v", err)
	}

	log.Println("worker started")
	<-ctx.Done()
	log.Println("worker stopped")
}
