package services

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type fakeAppRepoForAssign struct {
	app       *models.Application
	err       error
	updateErr error
}

func (r *fakeAppRepoForAssign) GetByID(id string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAppRepoForAssign) ListDueWithin(_, _ time.Time) ([]models.Application, error) {
	if r.app == nil {
		return nil, r.err
	}
	return []models.Application{*r.app}, r.err
}
func (r *fakeAppRepoForAssign) AttachmentExistsForApplication(_ string) (bool, error) {
	return false, nil
}
func (r *fakeAppRepoForAssign) UpdateAssignedStaff(_ string, _ *string, _ string) error {
	return r.updateErr
}

// embed other methods to satisfy interface
func (r *fakeAppRepoForAssign) CreateWithAttachments(app *models.Application, atts []models.ApplicationAttachment, notif *models.Notification, codeGen func() string) error {
	return nil
}
func (r *fakeAppRepoForAssign) ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error) {
	return nil, 0, nil
}
func (r *fakeAppRepoForAssign) ListStatusLogsByCitizen(appID, citizenUserID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (r *fakeAppRepoForAssign) CreateAttachments(appID string, atts []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeAppRepoForAssign) AdminList(_ repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	return nil, 0, nil
}
func (r *fakeAppRepoForAssign) GetByIDForCitizen(id, citizenUserID string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAppRepoForAssign) ProcessStatusUpdate(_ string, _ *models.ApplicationStatus, _ models.ApplicationStatus, _ string, _ string, _, _ *time.Time, _ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeAppRepoForAssign) GetDashboardStats() (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAppRepoForAssign) ListRecent(_ int) ([]models.Application, error) { return nil, nil }
func (r *fakeAppRepoForAssign) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAppRepoForAssign) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return nil, nil
}

type fakeAssignRepo struct{ created bool }

func (r *fakeAssignRepo) Create(a *models.ApplicationAssignment) (*models.ApplicationAssignment, error) {
	r.created = true
	a.ID = "a1"
	a.CreatedAt = time.Now()
	return a, nil
}
func (r *fakeAssignRepo) ListByApplication(applicationID string) ([]models.ApplicationAssignment, error) {
	return nil, nil
}

type fakeUserRepoAssign struct{ u *models.User }

func (r *fakeUserRepoAssign) FindByID(id string) (*models.User, error) {
	if r.u == nil {
		return nil, nil
	}
	return r.u, nil
}

// other user repo methods
func (r *fakeUserRepoAssign) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}
func (r *fakeUserRepoAssign) Create(user *models.User) (*models.User, error) {
	return nil, nil
}
func (r *fakeUserRepoAssign) CreateInTx(tx *gorm.DB, user *models.User) error {
	return nil
}
func (r *fakeUserRepoAssign) Update(user *models.User) error {
	return nil
}
func (r *fakeUserRepoAssign) List(filter repositories.UserFilter, offset, limit int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeUserRepoAssign) UpdateStatus(id string, status models.UserStatus, updatedBy string) error {
	return nil
}
func (r *fakeUserRepoAssign) SoftDelete(id string, deletedBy string) error {
	return nil
}

func TestAssignApplication_AssignAndRecord(t *testing.T) {
	app := &models.Application{ID: "app1", AssignedStaffUserID: nil}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1", Name: "Staff"}}
	logger := &fakeActivityLogger{}

	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo, logger)
	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.NoError(t, err)
	assert.True(t, assignRepo.created)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "assignment.assign", logger.entries[0].Action)
		assert.False(t, logger.entries[0].CreatedAt.IsZero())
		var metadata map[string]any
		assert.NoError(t, json.Unmarshal(logger.entries[0].MetadataJSON, &metadata))
		changes, ok := metadata["changes"].(map[string]any)
		assert.True(t, ok)
		assigned, ok := changes["assigned_staff_user_id"].(map[string]any)
		assert.True(t, ok)
		assert.Nil(t, assigned["before"])
		assert.Equal(t, "u1", assigned["after"])
	}
}

func TestAssignApplication_ApplicationNotFound(t *testing.T) {
	appRepo := &fakeAppRepoForAssign{app: nil, err: errors.New("not found")}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo)
	err := svc.AssignApplicationToStaff("appX", ptrStr("u1"), "admin-1")
	assert.Error(t, err)
}

func TestAssignApplication_Transfer(t *testing.T) {
	prevStaff := "u0"
	app := &models.Application{ID: "app1", AssignedStaffUserID: &prevStaff}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	logger := &fakeActivityLogger{}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo, logger)
	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.NoError(t, err)
	assert.True(t, assignRepo.created)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "assignment.transfer", logger.entries[0].Action)
	}
}

func TestAssignApplication_ApplicationNilWithoutError(t *testing.T) {
	appRepo := &fakeAppRepoForAssign{app: nil, err: nil}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo)
	err := svc.AssignApplicationToStaff("appX", ptrStr("u1"), "admin-1")
	assert.ErrorIs(t, err, ErrApplicationNotFoundAssign)
}

func TestAssignApplication_Unassign(t *testing.T) {
	prevStaff := "u0"
	app := &models.Application{ID: "app1", AssignedStaffUserID: &prevStaff}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{}
	logger := &fakeActivityLogger{}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo, logger)
	err := svc.AssignApplicationToStaff("app1", nil, "admin-1")
	assert.NoError(t, err)
	assert.True(t, assignRepo.created)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "assignment.unassign", logger.entries[0].Action)
	}
}

func TestAssignApplication_LogFailureDoesNotBreakMainFlow(t *testing.T) {
	app := &models.Application{ID: "app1", AssignedStaffUserID: nil}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1", Name: "Staff"}}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo, &fakeActivityLogger{err: errors.New("log failed")})

	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.NoError(t, err)
}

func TestAssignApplication_NoOp(t *testing.T) {
	sameID := "u1"
	app := &models.Application{ID: "app1", AssignedStaffUserID: &sameID}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo)
	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.NoError(t, err)
	assert.False(t, assignRepo.created)
}

func TestAssignApplication_UserNotFound(t *testing.T) {
	app := &models.Application{ID: "app1", AssignedStaffUserID: nil}
	appRepo := &fakeAppRepoForAssign{app: app}
	assignRepo := &fakeAssignRepo{}
	userRepo := &fakeUserRepoAssign{u: nil} // user not found
	svc := NewApplicationAssignmentService(appRepo, assignRepo, userRepo)
	err := svc.AssignApplicationToStaff("app1", ptrStr("missing"), "admin-1")
	assert.ErrorIs(t, err, ErrUserNotFoundAssign)
}

type fakeAssignRepoErr struct{}

func (r *fakeAssignRepoErr) Create(_ *models.ApplicationAssignment) (*models.ApplicationAssignment, error) {
	return nil, errors.New("create error")
}
func (r *fakeAssignRepoErr) ListByApplication(_ string) ([]models.ApplicationAssignment, error) {
	return nil, nil
}

func TestAssignApplication_CreateAssignmentError(t *testing.T) {
	app := &models.Application{ID: "app1", AssignedStaffUserID: nil}
	appRepo := &fakeAppRepoForAssign{app: app}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	svc := NewApplicationAssignmentService(appRepo, &fakeAssignRepoErr{}, userRepo)
	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.Error(t, err)
}

func TestAssignApplication_UpdateAssignedStaffError(t *testing.T) {
	app := &models.Application{ID: "app1", AssignedStaffUserID: nil}
	appRepo := &fakeAppRepoForAssign{app: app, updateErr: errors.New("update failed")}
	userRepo := &fakeUserRepoAssign{u: &models.User{ID: "u1"}}
	svc := NewApplicationAssignmentService(appRepo, &fakeAssignRepo{}, userRepo)
	err := svc.AssignApplicationToStaff("app1", ptrStr("u1"), "admin-1")
	assert.Error(t, err)
}

func ptrStr(s string) *string { return &s }
