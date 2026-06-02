package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"gorm.io/gorm"
)

const codeRetryAttempts = 3

type DashboardStats struct {
	Total    int64
	Pending  int64
	Approved int64
	Rejected int64
}

type DashboardRepository interface {
	GetDashboardStats() (DashboardStats, error)
	ListRecent(limit int) ([]models.Application, error)
	GetDashboardStatsForStaff(staffID string) (DashboardStats, error)
	ListRecentForStaff(staffID string, limit int) ([]models.Application, error)
}

type ApplicationRepository interface {
	CreateWithAttachments(app *models.Application, atts []models.ApplicationAttachment, notif *models.Notification, codeGen func() string) error
	GetByID(id string) (*models.Application, error)
	AdminList(filter ApplicationFilter, page, limit int) ([]models.Application, int64, error)
	ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error)
	GetByIDForCitizen(id, citizenUserID string) (*models.Application, error)
	ListStatusLogsByCitizen(appID, citizenUserID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error)
	CreateAttachments(appID string, atts []models.ApplicationAttachment) error
	UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error
	ProcessStatusUpdate(appID string, oldStatus *models.ApplicationStatus, newStatus models.ApplicationStatus, resultNote string, rejectedReason string, processingStartedAt, completedAt *time.Time, updatedBy string, atts []models.ApplicationAttachment) error
	DashboardRepository
}

type ApplicationFilter struct {
	Status              string
	Service             string
	Submitter           string
	AssignedStaffUserID string
}

type applicationRepo struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) CreateWithAttachments(
	app *models.Application,
	atts []models.ApplicationAttachment,
	notif *models.Notification,
	codeGen func() string,
) error {
	var lastErr error
	for attempt := 0; attempt < codeRetryAttempts; attempt++ {
		err := r.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(app).Error; err != nil {
				return err
			}
			for i := range atts {
				atts[i].ApplicationID = app.ID
			}
			if len(atts) > 0 {
				if err := tx.Create(&atts).Error; err != nil {
					return err
				}
			}
			notif.ApplicationID = &app.ID
			return tx.Create(notif).Error
		})
		if err == nil {
			return nil
		}
		lastErr = err
		if !isApplicationCodeConflict(err) {
			return err
		}
		app.ID = ""
		app.ApplicationCode = codeGen()
		for i := range atts {
			atts[i].ID = ""
			atts[i].ApplicationID = ""
		}
		notif.ID = ""
		notif.ApplicationID = nil
	}
	return lastErr
}

func (r *applicationRepo) GetByID(id string) (*models.Application, error) {
	var app models.Application
	if err := r.db.Preload("CitizenUser", "deleted_at IS NULL").
		Preload("ServiceType", "deleted_at IS NULL").
		Preload("ServiceType.ResponsibleDepartment", "deleted_at IS NULL").
		Preload("ServiceType.ResponsibleDepartment.LeaderUser", "deleted_at IS NULL").
		Preload("AssignedStaffUser", "deleted_at IS NULL").
		Preload("ApplicationAttachments", "deleted_at IS NULL").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepo) AdminList(filter ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	q := r.db.Model(&models.Application{}).
		Joins("LEFT JOIN service_types ON service_types.id = applications.service_type_id").
		Joins("LEFT JOIN users ON users.id = applications.citizen_user_id").
		Where("applications.deleted_at IS NULL").
		Where("service_types.deleted_at IS NULL").
		Where("users.deleted_at IS NULL")

	if filter.Status != "" {
		q = q.Where("applications.status = ?", filter.Status)
	}
	if filter.Service != "" {
		like := "%" + strings.ToLower(filter.Service) + "%"
		q = q.Where("LOWER(service_types.name) LIKE ? OR LOWER(service_types.code) LIKE ?", like, like)
	}
	if filter.Submitter != "" {
		like := "%" + strings.ToLower(filter.Submitter) + "%"
		q = q.Where("LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", like, like)
	}
	if filter.AssignedStaffUserID != "" {
		q = q.Where("applications.assigned_staff_user_id = ?", filter.AssignedStaffUserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.Application, 0)
	if err := q.Preload("ServiceType", "deleted_at IS NULL").
		Preload("CitizenUser", "deleted_at IS NULL").
		Preload("AssignedStaffUser", "deleted_at IS NULL").
		Order("submitted_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *applicationRepo) ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error) {
	q := r.db.Model(&models.Application{}).
		Where("citizen_user_id = ? AND deleted_at IS NULL", citizenUserID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.Application, 0)
	if err := q.Preload("ServiceType", "deleted_at IS NULL").
		Order("submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *applicationRepo) GetByIDForCitizen(id, citizenUserID string) (*models.Application, error) {
	var app models.Application
	if err := r.db.Preload("ServiceType", "deleted_at IS NULL").
		Preload("ApplicationAttachments", "deleted_at IS NULL").
		Where("id = ? AND citizen_user_id = ? AND deleted_at IS NULL", id, citizenUserID).
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepo) ListStatusLogsByCitizen(appID, citizenUserID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	q := r.db.Model(&models.ApplicationStatusLog{}).
		Joins("JOIN applications ON applications.id = application_status_logs.application_id").
		Where("application_status_logs.application_id = ? AND applications.citizen_user_id = ? AND applications.deleted_at IS NULL", appID, citizenUserID)

	if since != nil {
		q = q.Where("application_status_logs.created_at > ?", *since)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.ApplicationStatusLog, 0)
	if err := q.Preload("ChangedByUser").
		Order("application_status_logs.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *applicationRepo) CreateAttachments(appID string, atts []models.ApplicationAttachment) error {
	if len(atts) == 0 {
		return nil
	}
	for i := range atts {
		atts[i].ApplicationID = appID
	}
	return r.db.Create(&atts).Error
}

func (r *applicationRepo) UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error {
	return r.db.Model(&models.Application{}).
		Where("id = ? AND deleted_at IS NULL", applicationID).
		Updates(map[string]interface{}{
			"assigned_staff_user_id": assignedStaffUserID,
			"updated_at":             time.Now(),
		}).Error
}

func (r *applicationRepo) ProcessStatusUpdate(appID string, oldStatus *models.ApplicationStatus, newStatus models.ApplicationStatus, resultNote string, rejectedReason string, processingStartedAt, completedAt *time.Time, updatedBy string, atts []models.ApplicationAttachment) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status":                newStatus,
			"result_note":           resultNote,
			"rejected_reason":       rejectedReason,
			"processing_started_at": processingStartedAt,
			"completed_at":          completedAt,
			"updated_at":            now,
		}
		if err := tx.Model(&models.Application{}).Where("id = ? AND deleted_at IS NULL", appID).Updates(updates).Error; err != nil {
			return err
		}

		for i := range atts {
			atts[i].ApplicationID = appID
		}
		if len(atts) > 0 {
			if err := tx.Create(&atts).Error; err != nil {
				return err
			}
		}

		changedBy := updatedBy
		log := &models.ApplicationStatusLog{
			ApplicationID:   appID,
			OldStatus:       oldStatus,
			NewStatus:       newStatus,
			ChangedByUserID: &changedBy,
			Note:            resultNote,
			CreatedAt:       now,
		}
		return tx.Create(log).Error
	})
}

func isApplicationCodeConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "applications_application_code") || strings.Contains(msg, "application_code")
}

func (r *applicationRepo) GetDashboardStats() (DashboardStats, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.Model(&models.Application{}).
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&rows).Error; err != nil {
		return DashboardStats{}, err
	}
	var stats DashboardStats
	for _, row := range rows {
		stats.Total += row.Count
		switch models.ApplicationStatus(row.Status) {
		case models.ApplicationStatusReceived, models.ApplicationStatusProcessing, models.ApplicationStatusNeedMoreInfo:
			stats.Pending += row.Count
		case models.ApplicationStatusApproved:
			stats.Approved = row.Count
		case models.ApplicationStatusRejected:
			stats.Rejected = row.Count
		}
	}
	return stats, nil
}

func (r *applicationRepo) ListRecent(limit int) ([]models.Application, error) {
	items := make([]models.Application, 0, limit)
	if err := r.db.Preload("ServiceType", "deleted_at IS NULL").
		Preload("CitizenUser", "deleted_at IS NULL").
		Where("deleted_at IS NULL").
		Order("submitted_at DESC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *applicationRepo) GetDashboardStatsForStaff(staffID string) (DashboardStats, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.Model(&models.Application{}).
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL AND assigned_staff_user_id = ?", staffID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return DashboardStats{}, err
	}
	var stats DashboardStats
	for _, row := range rows {
		stats.Total += row.Count
		switch models.ApplicationStatus(row.Status) {
		case models.ApplicationStatusReceived, models.ApplicationStatusProcessing, models.ApplicationStatusNeedMoreInfo:
			stats.Pending += row.Count
		case models.ApplicationStatusApproved:
			stats.Approved = row.Count
		case models.ApplicationStatusRejected:
			stats.Rejected = row.Count
		}
	}
	return stats, nil
}

func (r *applicationRepo) ListRecentForStaff(staffID string, limit int) ([]models.Application, error) {
	items := make([]models.Application, 0, limit)
	if err := r.db.Preload("ServiceType", "deleted_at IS NULL").
		Preload("CitizenUser", "deleted_at IS NULL").
		Where("deleted_at IS NULL AND assigned_staff_user_id = ?", staffID).
		Order("submitted_at DESC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

var _ func() string = utils.GenerateApplicationCode
