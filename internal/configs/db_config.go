package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("database configuration error: DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get sql.DB: %v", err))
	}
	// Recycle connections every 30 min so they don't outlive PostgreSQL's
	// server-side idle timeout, avoiding stale-connection errors.
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		panic(fmt.Sprintf("failed to enable pgcrypto extension: %v", err))
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.CitizenProfile{},
		&models.Department{},
		&models.Category{},
		&models.StaffProfile{},
		&models.ServiceType{},
		&models.Application{},
		&models.ApplicationAttachment{},
		&models.ApplicationStatusLog{},
		&models.ApplicationAssignment{},
		&models.ApplicationReminderLog{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.ImportExportLog{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	return db
}
