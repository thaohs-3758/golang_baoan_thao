package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/consumers"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/queue"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
)

func main() {
	_ = godotenv.Load()

	if err := configs.LoadI18nMessages("locales"); err != nil {
		log.Fatalf("failed to load i18n: %v", err)
	}

	db := configs.InitDB()
	repo := repositories.NewNotificationRepository(db)
	eventSvc := services.NewNotificationEventService(repo)
	eventConsumer := consumers.NewNotificationEventConsumer(eventSvc)
	apiSvc := services.NewNotificationAPIService(repo)
	apiHandler := handlers.NewNotificationServiceHandler(apiSvc)

	rabbitConn, err := queue.NewRabbitMQConnection()
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer rabbitConn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	consumer := queue.NewRabbitMQConsumer(rabbitConn)
	if err := consumer.ConsumeApplicationSubmitted(ctx, "notification.application.submitted", eventConsumer.HandleApplicationSubmitted); err != nil {
		log.Fatalf("failed to start application submitted consumer: %v", err)
	}
	if err := consumer.ConsumeApplicationStatusChanged(ctx, "notification.application.status_changed", eventConsumer.HandleApplicationStatusChanged); err != nil {
		log.Fatalf("failed to start application status changed consumer: %v", err)
	}
	if err := consumer.ConsumeApplicationDeadlineReminder(ctx, "notification.application.deadline_reminder", eventConsumer.HandleApplicationDeadlineReminder); err != nil {
		log.Fatalf("failed to start application deadline reminder consumer: %v", err)
	}

	e := echo.New()
	e.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/notifications", apiHandler.List)
	e.GET("/notifications/unread-count", apiHandler.CountUnread)
	e.PUT("/notifications/:id/read", apiHandler.MarkAsRead)
	e.PUT("/notifications/read-all", apiHandler.MarkAllAsRead)

	log.Printf("notification-service listening on %s", notificationServiceAddr())
	sc := echo.StartConfig{
		Address:         notificationServiceAddr(),
		GracefulTimeout: shutdownTimeout(),
	}
	if err := sc.Start(ctx, e); err != nil {
		e.Logger.Error("failed to start notification-service", "error", err)
	}
}

func notificationServiceAddr() string {
	if port := os.Getenv("NOTIFICATION_SERVICE_PORT"); port != "" {
		return ":" + port
	}
	return ":8081"
}

func shutdownTimeout() time.Duration {
	return 5 * time.Second
}
