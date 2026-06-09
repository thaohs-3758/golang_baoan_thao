package models

import "time"

type ApplicationReminderLog struct {
	ID              string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationID   string     `json:"application_id" gorm:"type:uuid;not null;index:idx_app_reminder_unique,unique"`
	RecipientUserID string     `json:"recipient_user_id" gorm:"type:uuid;not null;index:idx_app_reminder_unique,unique"`
	ReminderType    string     `json:"reminder_type" gorm:"type:varchar(50);not null;index:idx_app_reminder_unique,unique"`
	SentAt          time.Time  `json:"sent_at" gorm:"not null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"not null"`
	DeletedAt       *time.Time `json:"deleted_at" gorm:"index"`
}
