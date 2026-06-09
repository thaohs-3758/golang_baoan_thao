package models

import (
	"encoding/json"
	"time"
)

type Application struct {
	ID                  string            `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationCode     string            `json:"application_code" gorm:"type:varchar(100);not null;uniqueIndex"`
	CitizenUserID       string            `json:"citizen_user_id" gorm:"type:uuid;not null;index"`
	CitizenUser         User              `json:"citizen_user" gorm:"foreignKey:CitizenUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ServiceTypeID       string            `json:"service_type_id" gorm:"type:uuid;not null;index"`
	ServiceType         ServiceType       `json:"service_type" gorm:"foreignKey:ServiceTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	AssignedStaffUserID *string           `json:"assigned_staff_user_id" gorm:"type:uuid;index"`
	AssignedStaffUser   *User             `json:"assigned_staff_user" gorm:"foreignKey:AssignedStaffUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Status              ApplicationStatus `json:"status" gorm:"type:varchar(30);not null;default:'received';index;check:status IN ('received','processing','need_more_info','approved','rejected')"`
	SubmittedData       json.RawMessage   `json:"submitted_data" gorm:"type:jsonb;not null"`
	ResultNote          string            `json:"result_note" gorm:"type:text"`
	RejectedReason      string            `json:"rejected_reason" gorm:"type:text"`
	SubmittedAt         time.Time         `json:"submitted_at" gorm:"not null;index"`
	DueAt               *time.Time        `json:"due_at" gorm:"index"`
	ProcessingStartedAt *time.Time        `json:"processing_started_at"`
	CompletedAt         *time.Time        `json:"completed_at"`
	CreatedAt           time.Time         `json:"created_at" gorm:"not null"`
	UpdatedAt           time.Time         `json:"updated_at" gorm:"not null"`
	DeletedAt           *time.Time        `json:"deleted_at" gorm:"index"`

	ApplicationAttachments []ApplicationAttachment `json:"application_attachments,omitempty" gorm:"foreignKey:ApplicationID"`
}

// SubmittedAtFormatted returns SubmittedAt formatted as YYYY-MM-DD HH:mm
func (a Application) SubmittedAtFormatted() string {
	if a.SubmittedAt.IsZero() {
		return ""
	}
	return a.SubmittedAt.Format("2006-01-02 15:04")
}
