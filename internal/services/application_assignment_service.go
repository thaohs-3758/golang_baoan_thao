package services

import (
	"errors"
	"log"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrApplicationNotFoundAssign = errors.New("application.not_found")
var ErrUserNotFoundAssign = errors.New("user.not_found")

type ApplicationAssignmentService struct {
	appRepo        repositories.ApplicationRepository
	assignRepo     repositories.ApplicationAssignmentRepository
	userRepo       repositories.UserRepository
	activityLogger activityLogger
}

func NewApplicationAssignmentService(appRepo repositories.ApplicationRepository, assignRepo repositories.ApplicationAssignmentRepository, userRepo repositories.UserRepository, loggers ...activityLogger) *ApplicationAssignmentService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &ApplicationAssignmentService{appRepo: appRepo, assignRepo: assignRepo, userRepo: userRepo, activityLogger: logger}
}

func (s *ApplicationAssignmentService) AssignApplicationToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil || app == nil {
		return ErrApplicationNotFoundAssign
	}

	var toUser *models.User
	if toStaffUserID != nil && *toStaffUserID != "" {
		toUser, err = s.userRepo.FindByID(*toStaffUserID)
		if err != nil || toUser == nil {
			return ErrUserNotFoundAssign
		}
	}

	var action models.AssignmentAction
	if app.AssignedStaffUserID == nil && toStaffUserID != nil {
		action = models.AssignmentActionAssigned
	} else if app.AssignedStaffUserID != nil && toStaffUserID != nil && *app.AssignedStaffUserID != *toStaffUserID {
		action = models.AssignmentActionTransferred
	} else if toStaffUserID == nil {
		action = models.AssignmentActionUnassigned
	} else {
		// no-op (assigning to same user)
		return nil
	}

	assignment := &models.ApplicationAssignment{
		ApplicationID:    applicationID,
		FromStaffUserID:  app.AssignedStaffUserID,
		ToStaffUserID:    toStaffUserID,
		AssignedByUserID: &assignedBy,
		Action:           action,
		Note:             "",
	}

	if _, err := s.assignRepo.Create(assignment); err != nil {
		return err
	}

	if err := s.appRepo.UpdateAssignedStaff(applicationID, toStaffUserID, assignedBy); err != nil {
		return err
	}
	s.logAssignmentActivity(applicationID, app.ApplicationCode, assignedBy, action, app.AssignedStaffUserID, toStaffUserID)

	return nil
}

func (s *ApplicationAssignmentService) logAssignmentActivity(applicationID, applicationCode, assignedBy string, action models.AssignmentAction, fromStaffUserID, toStaffUserID *string) {
	if s.activityLogger == nil {
		return
	}
	var actionName string
	switch action {
	case models.AssignmentActionAssigned:
		actionName = "assignment.assign"
	case models.AssignmentActionTransferred:
		actionName = "assignment.transfer"
	case models.AssignmentActionUnassigned:
		actionName = "assignment.unassign"
	default:
		return
	}
	actorID := assignedBy
	description := "Phân công hồ sơ " + applicationCode
	entry := &models.ActivityLog{
		ActorUserID: &actorID,
		Action:      actionName,
		EntityType:  "application",
		EntityID:    &applicationID,
		Description: description,
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"changes": map[string]any{
				"assigned_staff_user_id": map[string]any{
					"before": fromStaffUserID,
					"after":  toStaffUserID,
				},
			},
		}),
		CreatedAt: time.Now(),
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}
