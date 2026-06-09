package events

import "time"

const ApplicationDeadlineReminder = "application.deadline_reminder"

type ApplicationDeadlineReminderEvent struct {
	ApplicationID   string    `json:"application_id"`
	ApplicationCode string    `json:"application_code"`
	RecipientUserID string    `json:"recipient_user_id"`
	RecipientRole   string    `json:"recipient_role"`
	ServiceName     string    `json:"service_name"`
	DueAt           time.Time `json:"due_at"`
	ReminderType    string    `json:"reminder_type"`
}
