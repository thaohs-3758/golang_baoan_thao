package models

import "time"

type Notification struct {
	ID            string           `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        string           `json:"user_id" gorm:"type:uuid;not null;index"`
	User          User             `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApplicationID *string          `json:"application_id" gorm:"type:uuid;index"`
	Application   *Application     `json:"application" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	SourceEventID *string          `json:"source_event_id" gorm:"type:varchar(100);uniqueIndex"`
	Title         string           `json:"title" gorm:"type:varchar(255);not null"`
	Message       string           `json:"message" gorm:"type:text;not null"`
	Type          NotificationType `json:"type" gorm:"type:varchar(30);not null;default:'system';check:type IN ('received','need_more_info','result','system','deadline_reminder')"`
	IsRead        bool             `json:"is_read" gorm:"not null;default:false;index"`
	ReadAt        *time.Time       `json:"read_at"`
	CreatedAt     time.Time        `json:"created_at" gorm:"not null"`
	DeletedAt     *time.Time       `json:"deleted_at" gorm:"index"`
}
