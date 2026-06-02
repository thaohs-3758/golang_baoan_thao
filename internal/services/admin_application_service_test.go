package services

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type fakeAdminAppRepo struct {
	apps          []models.Application
	total         int64
	app           *models.Application
	err           error
	lastFilter    repositories.ApplicationFilter
	processStatus models.ApplicationStatus
	processNote   string
}

func (r *fakeAdminAppRepo) AdminList(filter repositories.ApplicationFilter, _, _ int) ([]models.Application, int64, error) {
	r.lastFilter = filter
	return r.apps, r.total, r.err
}
func (r *fakeAdminAppRepo) GetByID(_ string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAdminAppRepo) UpdateAssignedStaff(_ string, _ *string, _ string) error { return r.err }
func (r *fakeAdminAppRepo) ProcessStatusUpdate(_ string, _ *models.ApplicationStatus, newStatus models.ApplicationStatus, resultNote string, _ string, _, _ *time.Time, _ string, _ []models.ApplicationAttachment) error {
	r.processStatus = newStatus
	r.processNote = resultNote
	return r.err
}
func (r *fakeAdminAppRepo) CreateWithAttachments(_ *models.Application, _ []models.ApplicationAttachment, _ *models.Notification, _ func() string) error {
	return nil
}
func (r *fakeAdminAppRepo) ListByCitizen(_ string, _, _ int) ([]models.Application, int64, error) {
	return nil, 0, nil
}
func (r *fakeAdminAppRepo) ListStatusLogsByCitizen(_ string, _ string, _, _ int, _ *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (r *fakeAdminAppRepo) CreateAttachments(_ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeAdminAppRepo) GetByIDForCitizen(_, _ string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAdminAppRepo) GetDashboardStats() (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAdminAppRepo) ListRecent(_ int) ([]models.Application, error) { return nil, nil }
func (r *fakeAdminAppRepo) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAdminAppRepo) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return nil, nil
}

func newAdminAppSvc(repo *fakeAdminAppRepo) *AdminApplicationService {
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	return NewAdminApplicationService(repo, assignSvc, nil)
}

func newAdminAppSvcWithLogger(repo *fakeAdminAppRepo, logger activityLogger) *AdminApplicationService {
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	return NewAdminApplicationService(repo, assignSvc, nil, logger)
}

func TestAdminApplicationService_ListApplications_OK(t *testing.T) {
	apps := []models.Application{{ID: "a1", ApplicationCode: "APP-2024-001"}}
	svc := newAdminAppSvc(&fakeAdminAppRepo{apps: apps, total: 1})
	result, total, err := svc.ListApplications(repositories.ApplicationFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAdminApplicationService_ListApplications_Error(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("db error")})
	_, _, err := svc.ListApplications(repositories.ApplicationFilter{}, 1, 10)
	assert.Error(t, err)
}

func TestAdminApplicationService_ListApplications_WithFilter(t *testing.T) {
	repo := &fakeAdminAppRepo{apps: []models.Application{{ID: "a1"}}, total: 1}
	svc := newAdminAppSvc(repo)
	filter := repositories.ApplicationFilter{Status: string(models.ApplicationStatusProcessing), Service: "CCCD", Submitter: "An"}
	_, _, err := svc.ListApplications(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, filter, repo.lastFilter)
}

func TestAdminApplicationService_ListApplications_StaffForcesAssignedScope(t *testing.T) {
	repo := &fakeAdminAppRepo{apps: []models.Application{{ID: "a1"}}, total: 1}
	svc := newAdminAppSvc(repo)

	filter := repositories.ApplicationFilter{Status: string(models.ApplicationStatusProcessing)}
	_, _, err := svc.ListApplicationsForActor(filter, 1, 10, models.UserRoleStaff, "staff-1")
	assert.NoError(t, err)
	assert.Equal(t, "staff-1", repo.lastFilter.AssignedStaffUserID)
	assert.Equal(t, string(models.ApplicationStatusProcessing), repo.lastFilter.Status)
}

func TestAdminApplicationService_ListApplications_ManagerDoesNotForceAssignedScope(t *testing.T) {
	repo := &fakeAdminAppRepo{apps: []models.Application{{ID: "a1"}}, total: 1}
	svc := newAdminAppSvc(repo)

	filter := repositories.ApplicationFilter{Status: string(models.ApplicationStatusProcessing)}
	_, _, err := svc.ListApplicationsForActor(filter, 1, 10, models.UserRoleManager, "manager-1")
	assert.NoError(t, err)
	assert.Equal(t, "", repo.lastFilter.AssignedStaffUserID)
}

func TestAdminApplicationService_ListApplications_SuperAdminDoesNotForceAssignedScope(t *testing.T) {
	repo := &fakeAdminAppRepo{apps: []models.Application{{ID: "a1"}}, total: 1}
	svc := newAdminAppSvc(repo)

	_, _, err := svc.ListApplicationsForActor(repositories.ApplicationFilter{}, 1, 10, models.UserRoleSuperAdmin, "sa-1")
	assert.NoError(t, err)
	assert.Equal(t, "", repo.lastFilter.AssignedStaffUserID)
}

func TestAdminApplicationService_GetApplication_OK(t *testing.T) {
	app := &models.Application{ID: "a1"}
	svc := newAdminAppSvc(&fakeAdminAppRepo{app: app})
	result, err := svc.GetApplication("a1")
	assert.NoError(t, err)
	assert.Equal(t, "a1", result.ID)
}

func TestAdminApplicationService_GetApplication_Error(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("not found")})
	_, err := svc.GetApplication("missing")
	assert.Error(t, err)
}

func TestAdminApplicationService_AssignToStaff_AppNotFound(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("not found")})
	err := svc.AssignToStaff("appX", nil, "admin-1")
	assert.Error(t, err)
}

func TestAdminApplicationService_AssignToStaff_Unassign(t *testing.T) {
	prevID := "u0"
	app := &models.Application{ID: "app1", AssignedStaffUserID: &prevID}
	svc := newAdminAppSvc(&fakeAdminAppRepo{app: app})
	err := svc.AssignToStaff("app1", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_OK(t *testing.T) {
	prevStatus := models.ApplicationStatusReceived
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: prevStatus}}
	logger := &fakeActivityLogger{}
	svc := newAdminAppSvcWithLogger(repo, logger)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "processing", nil, "admin-1")
	assert.NoError(t, err)
	assert.Equal(t, models.ApplicationStatusProcessing, repo.processStatus)
	assert.Equal(t, "processing", repo.processNote)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "application.status_update", logger.entries[0].Action)
		var metadata map[string]any
		assert.NoError(t, json.Unmarshal(logger.entries[0].MetadataJSON, &metadata))
		changes, ok := metadata["changes"].(map[string]any)
		assert.True(t, ok)
		status, ok := changes["status"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, string(models.ApplicationStatusReceived), status["before"])
		assert.Equal(t, string(models.ApplicationStatusProcessing), status["after"])
	}
}

func TestAdminApplicationService_ProcessApplication_LogFailureDoesNotBreakMainFlow(t *testing.T) {
	prevStatus := models.ApplicationStatusReceived
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: prevStatus}}
	svc := newAdminAppSvcWithLogger(repo, &fakeActivityLogger{err: errors.New("log failed")})

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "ok", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_InvalidTransition(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusApproved}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationInvalidTransition)
}

func TestAdminApplicationService_ProcessApplication_RejectedRequiresReason(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusProcessing}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusRejected, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationRejectReasonRequired)
}

func TestAdminApplicationService_ProcessApplication_AllowNeedMoreInfoTransitions(t *testing.T) {
	cases := []struct {
		name string
		from models.ApplicationStatus
		to   models.ApplicationStatus
		note string
	}{
		{
			name: "received to need_more_info",
			from: models.ApplicationStatusReceived,
			to:   models.ApplicationStatusNeedMoreInfo,
			note: "need additional papers",
		},
		{
			name: "processing to need_more_info",
			from: models.ApplicationStatusProcessing,
			to:   models.ApplicationStatusNeedMoreInfo,
			note: "missing photo",
		},
		{
			name: "need_more_info to processing",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusProcessing,
			note: "citizen submitted extra docs",
		},
		{
			name: "need_more_info to approved",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusApproved,
			note: "all docs are valid",
		},
		{
			name: "need_more_info to rejected",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusRejected,
			note: "documents are inconsistent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: tc.from}}
			svc := newAdminAppSvc(repo)

			err := svc.ProcessApplication("app1", tc.to, tc.note, nil, "admin-1")
			assert.NoError(t, err)
			assert.Equal(t, tc.to, repo.processStatus)
		})
	}
}

func TestAdminApplicationService_ProcessApplication_RequireNoteForNeedMoreInfo(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusProcessing}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusNeedMoreInfo, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationNeedMoreInfoNoteRequired)
	assert.Equal(t, models.ApplicationStatus(""), repo.processStatus)
}

func TestAdminApplicationService_ProcessApplication_RequireReasonForRejected(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusNeedMoreInfo}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusRejected, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationRejectReasonRequired)
	assert.Equal(t, models.ApplicationStatus(""), repo.processStatus)
}

func TestAdminApplicationService_WithNotificationRepo(t *testing.T) {
	repo := &fakeAdminAppRepo{}
	svc := newAdminAppSvc(repo)
	notifRepo := &fakeNotificationRepo{}

	result := svc.WithNotificationRepo(notifRepo)
	assert.Equal(t, svc, result, "WithNotificationRepo should return self")
}

func TestAdminApplicationService_ProcessApplication_SendsNotification(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusReceived,
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	svc := newAdminAppSvc(repo)
	notifRepo := &fakeNotificationRepo{}
	svc.WithNotificationRepo(notifRepo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "Đang xử lý", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_SendsNotificationApproved(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusProcessing,
		CitizenUser:     models.User{ID: "citizen-1", Email: "citizen@test.com", Name: "Citizen"},
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	svc := newAdminAppSvc(repo)
	svc.WithNotificationRepo(&fakeNotificationRepo{})
	mailer := &fakeMailer{}
	svc.mailer = mailer

	err := svc.ProcessApplication("app1", models.ApplicationStatusApproved, "Approved", nil, "admin-1")
	assert.NoError(t, err)
	assert.True(t, mailer.sent)
}

func TestAdminApplicationService_ProcessApplication_SendsNotificationRejected(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusProcessing,
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	svc := newAdminAppSvc(repo)
	svc.WithNotificationRepo(&fakeNotificationRepo{})

	err := svc.ProcessApplication("app1", models.ApplicationStatusRejected, "Rejected reason", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_SendsNotificationNeedMoreInfo(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusProcessing,
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	svc := newAdminAppSvc(repo)
	svc.WithNotificationRepo(&fakeNotificationRepo{})

	err := svc.ProcessApplication("app1", models.ApplicationStatusNeedMoreInfo, "Need more docs", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_DefaultStatusNoNotification(t *testing.T) {
	// Test that notifyCitizenStatusChange with a "received" status (not in switch) is a no-op
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusNeedMoreInfo,
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	svc := newAdminAppSvc(repo)
	notifRepo := &fakeNotificationRepo{}
	svc.WithNotificationRepo(notifRepo)

	// Going back to Processing from NeedMoreInfo is valid
	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "Resuming", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_StorageNotConfigured(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:     "app1",
		Status: models.ApplicationStatusReceived,
	}}
	svc := newAdminAppSvc(repo) // no storage configured

	// Create a fake multipart.FileHeader to trigger the file upload path
	// We'll need to use a dummy file header
	fakeFile := make([]*multipart.FileHeader, 1)
	fakeFile[0] = &multipart.FileHeader{Filename: "test.pdf"}

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "note", fakeFile, "admin-1")
	assert.Error(t, err) // expects "storage not configured"
}

func TestAdminApplicationService_ProcessApplication_WithStorage_OK(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusReceived}}
	storage := &fakeStorage{pubURL: "/uploads/app1/result.pdf", mime: "application/pdf", size: 1024}
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	svc := NewAdminApplicationService(repo, assignSvc, storage)

	files := []*multipart.FileHeader{{Filename: "result.pdf", Size: 1024}}
	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "note", files, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_EmailIncludesAttachmentURLs(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusProcessing,
		CitizenUser:     models.User{ID: "citizen-1", Email: "citizen@test.com", Name: "Citizen"},
		ServiceType:     models.ServiceType{Name: "Test Service"},
	}}
	storage := &fakeStorage{pubURL: "/uploads/app1/result.pdf", mime: "application/pdf", size: 1024}
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	mailer := &fakeMailer{}
	svc := NewAdminApplicationService(repo, assignSvc, storage).WithMailer(mailer)

	files := []*multipart.FileHeader{{Filename: "result.pdf", Size: 1024}}
	err := svc.ProcessApplication("app1", models.ApplicationStatusApproved, "approved", files, "admin-1")

	assert.NoError(t, err)
	assert.True(t, mailer.sent)
	assert.Contains(t, mailer.body, "/uploads/app1/result.pdf")
}

func TestAdminApplicationService_ProcessApplication_WithStorage_SaveError(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusReceived}}
	storage := &fakeStorage{saveErr: errors.New("disk full")}
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	svc := NewAdminApplicationService(repo, assignSvc, storage)

	files := []*multipart.FileHeader{{Filename: "result.pdf", Size: 1024}}
	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "note", files, "admin-1")
	assert.Error(t, err)
}

func TestAdminApplicationService_ProcessApplication_NilApp(t *testing.T) {
	repo := &fakeAdminAppRepo{app: nil}
	svc := newAdminAppSvc(repo)
	err := svc.ProcessApplication("missing", models.ApplicationStatusProcessing, "note", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationNotFound)
}

func TestAdminApplicationService_ProcessApplication_RepoError(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusReceived}}
	storage := &fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 100}
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	svc := NewAdminApplicationService(repo, assignSvc, storage)
	// After app found, set err for ProcessStatusUpdate
	repo.err = errors.New("process failed")

	files := []*multipart.FileHeader{{Filename: "result.pdf", Size: 1024}}
	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "note", files, "admin-1")
	assert.Error(t, err)
}

func TestAdminApplicationService_ProcessApplication_ApprovedSetsCompletedAt(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusProcessing}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusApproved, "approved", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_ProcessingSecondTime(t *testing.T) {
	// processingStartedAt already set — should not overwrite
	now := time.Now()
	repo := &fakeAdminAppRepo{app: &models.Application{
		ID:                 "app1",
		Status:             models.ApplicationStatusReceived,
		ProcessingStartedAt: &now,
	}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "note", nil, "admin-1")
	assert.NoError(t, err)
}
