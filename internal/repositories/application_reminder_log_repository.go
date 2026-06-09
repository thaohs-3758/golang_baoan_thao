package repositories

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type ApplicationReminderLogRepository interface {
	Exists(applicationID, recipientUserID, reminderType string) (bool, error)
	Create(log *models.ApplicationReminderLog) error
}

type applicationReminderLogRepo struct {
	db *gorm.DB
}

func NewApplicationReminderLogRepository(db *gorm.DB) ApplicationReminderLogRepository {
	return &applicationReminderLogRepo{db: db}
}

func (r *applicationReminderLogRepo) Exists(applicationID, recipientUserID, reminderType string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ApplicationReminderLog{}).
		Where("application_id = ? AND recipient_user_id = ? AND reminder_type = ? AND deleted_at IS NULL", applicationID, recipientUserID, reminderType).
		Count(&count).Error
	return count > 0, err
}

func (r *applicationReminderLogRepo) Create(log *models.ApplicationReminderLog) error {
	return r.db.Create(log).Error
}
