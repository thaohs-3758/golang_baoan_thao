package events

import "time"

const (
	ApplicationSubmitted     = "application.submitted"
	ApplicationStatusChanged = "application.status_changed"
)

type ApplicationSubmittedEvent struct {
	EventID         string    `json:"event_id"`
	ApplicationID   string    `json:"application_id"`
	ApplicationCode string    `json:"application_code"`
	CitizenUserID   string    `json:"citizen_user_id"`
	ServiceName     string    `json:"service_name"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type ApplicationStatusChangedEvent struct {
	EventID         string    `json:"event_id"`
	ApplicationID   string    `json:"application_id"`
	ApplicationCode string    `json:"application_code"`
	CitizenUserID   string    `json:"citizen_user_id"`
	ServiceName     string    `json:"service_name"`
	OldStatus       string    `json:"old_status"`
	NewStatus       string    `json:"new_status"`
	Note            string    `json:"note"`
	ChangedBy       string    `json:"changed_by"`
	OccurredAt      time.Time `json:"occurred_at"`
}
