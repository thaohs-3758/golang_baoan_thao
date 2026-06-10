package services

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// --- fakes ---

type fakeAppServiceTypeRepo struct {
	st  *models.ServiceType
	err error
}

func (r *fakeAppServiceTypeRepo) List(ctx context.Context, _ repositories.ListFilter) (*repositories.ListResult, error) {
	return nil, nil
}
func (r *fakeAppServiceTypeRepo) GetByID(ctx context.Context, _ string) (*models.ServiceType, error) {
	return r.st, r.err
}
func (r *fakeAppServiceTypeRepo) GetByIDForAdmin(ctx context.Context, _ string) (*models.ServiceType, error) {
	return r.st, r.err
}
func (r *fakeAppServiceTypeRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	return nil, nil
}
func (r *fakeAppServiceTypeRepo) ListDepartments(ctx context.Context) ([]models.Department, error) {
	return nil, nil
}
func (r *fakeAppServiceTypeRepo) Create(ctx context.Context, _ *models.ServiceType) error { return nil }
func (r *fakeAppServiceTypeRepo) CreateInTx(_ *gorm.DB, _ *models.ServiceType) error      { return nil }
func (r *fakeAppServiceTypeRepo) Update(ctx context.Context, _ *models.ServiceType) error { return nil }
func (r *fakeAppServiceTypeRepo) Delete(ctx context.Context, _ string) error              { return nil }
func (r *fakeAppServiceTypeRepo) CountApplications(ctx context.Context, _ string) (int64, error) {
	return 0, nil
}

type fakeAppUserRepo struct {
	user *models.User
	err  error
}

func (r *fakeAppUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeAppUserRepo) FindByID(_ string) (*models.User, error)     { return r.user, r.err }
func (r *fakeAppUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeAppUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return nil }
func (r *fakeAppUserRepo) Update(_ *models.User) error                 { return nil }
func (r *fakeAppUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeAppUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeAppUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

type fakeAppRepo struct {
	createErr    error
	createdApp   *models.Application
	apps         []models.Application
	total        int64
	listErr      error
	app          *models.Application
	getErr       error
	logs         []models.ApplicationStatusLog
	logsTotal    int64
	logsErr      error
	createAttErr error
}

func (r *fakeAppRepo) CreateWithAttachments(app *models.Application, _ []models.ApplicationAttachment, _ func() string) error {
	r.createdApp = app
	return r.createErr
}
func (r *fakeAppRepo) ListByCitizen(_ string, _, _ int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.listErr
}
func (r *fakeAppRepo) GetByIDForCitizen(_, _ string) (*models.Application, error) {
	return r.app, r.getErr
}
func (r *fakeAppRepo) ListStatusLogsByCitizen(_, _ string, _, _ int, _ *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return r.logs, r.logsTotal, r.logsErr
}
func (r *fakeAppRepo) CreateAttachments(_ string, _ []models.ApplicationAttachment) error {
	return r.createAttErr
}
func (r *fakeAppRepo) AdminList(_ repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.listErr
}
func (r *fakeAppRepo) GetByID(id string) (*models.Application, error) {
	if r.app != nil && r.app.ID == id {
		return r.app, r.getErr
	}
	return nil, r.getErr
}
func (r *fakeAppRepo) ListDueWithin(_, _ time.Time) ([]models.Application, error) {
	return r.apps, r.listErr
}
func (r *fakeAppRepo) AttachmentExistsForApplication(_ string) (bool, error) {
	return false, nil
}
func (r *fakeAppRepo) UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error {
	return nil
}
func (r *fakeAppRepo) ProcessStatusUpdate(_ string, _ *models.ApplicationStatus, _ models.ApplicationStatus, _ string, _ string, _, _ *time.Time, _ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeAppRepo) GetDashboardStats() (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAppRepo) ListRecent(_ int) ([]models.Application, error) { return nil, nil }
func (r *fakeAppRepo) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAppRepo) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return nil, nil
}

type fakeStorage struct {
	pubURL  string
	mime    string
	size    int64
	saveErr error
}

func (s *fakeStorage) SaveApplicationFile(_ string, fh *multipart.FileHeader) (string, string, int64, error) {
	if s.saveErr != nil {
		return "", "", 0, s.saveErr
	}
	return s.pubURL, s.mime, s.size, nil
}
func (s *fakeStorage) RemoveApplicationDir(_ string) error { return nil }
func (s *fakeStorage) RemoveFile(_ string) error           { return nil }
func (s *fakeStorage) ListApplicationDirs() ([]utils.TempDirInfo, error) {
	return nil, nil
}

type fakeMailer struct {
	sent bool
	body string
}

func (m *fakeMailer) Send(_, _, _, body string, _ ...string) error {
	m.sent = true
	m.body = body
	return nil
}

type fakeActivityLogger struct {
	entries []*models.ActivityLog
	err     error
}

func (l *fakeActivityLogger) Log(entry *models.ActivityLog) error {
	l.entries = append(l.entries, entry)
	return l.err
}

// compile-time interface checks
var _ repositories.ServiceTypeRepository = (*fakeAppServiceTypeRepo)(nil)
var _ repositories.UserRepository = (*fakeAppUserRepo)(nil)
var _ repositories.ApplicationRepository = (*fakeAppRepo)(nil)
var _ utils.FileStorage = (*fakeStorage)(nil)
var _ Mailer = (*fakeMailer)(nil)

// --- helpers ---

func makeSchema(required []string) json.RawMessage {
	b, _ := json.Marshal(map[string]interface{}{"required": required})
	return b
}

func makeData(fields map[string]string) json.RawMessage {
	b, _ := json.Marshal(fields)
	return b
}

func newSvc(
	appRepo *fakeAppRepo,
	stRepo *fakeAppServiceTypeRepo,
	uRepo *fakeAppUserRepo,
	storage *fakeStorage,
	mailer *fakeMailer,
	logger ...activityLogger,
) *ApplicationService {
	return NewApplicationService(appRepo, stRepo, uRepo, nil, storage, mailer, logger...)
}

func activeServiceType() *models.ServiceType {
	return &models.ServiceType{
		ID:         "st-1",
		Name:       "Cấp CCCD lần đầu",
		IsActive:   true,
		FormSchema: makeSchema([]string{"full_name", "date_of_birth"}),
	}
}

func validReq() *dtos.SubmitApplicationRequest {
	return &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{
			"full_name":     "Nguyen Van A",
			"date_of_birth": "1995-01-01",
		}),
	}
}

// --- SubmitApplication ---

func TestSubmitApplication_Success_NoFiles(t *testing.T) {
	logger := &fakeActivityLogger{}
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 100},
		&fakeMailer{},
		logger,
	)

	resp, err := svc.SubmitApplication("u1", validReq(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ApplicationCode)
	assert.Equal(t, "received", resp.Status)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "application.submit", logger.entries[0].Action)
		var metadata map[string]any
		assert.NoError(t, json.Unmarshal(logger.entries[0].MetadataJSON, &metadata))
		assert.Equal(t, "st-1", metadata["service_type_id"])
		assert.Equal(t, float64(0), metadata["attachments"])
	}
}

func TestSubmitApplication_SetsDueAtFromProcessingTime(t *testing.T) {
	processingDays := 2
	serviceType := &models.ServiceType{
		ID:             "svc-1",
		Name:           "Cap CCCD",
		IsActive:       true,
		FormSchema:     []byte(`{}`),
		ProcessingTime: &processingDays,
	}

	repo := &fakeAppRepo{}
	svc := newSvc(
		repo,
		&fakeAppServiceTypeRepo{st: serviceType},
		&fakeAppUserRepo{user: &models.User{ID: "citizen-1", Email: "a@b.com", Name: "Citizen"}},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("citizen-1", &dtos.SubmitApplicationRequest{
		ServiceTypeID: "svc-1",
		SubmittedData: []byte(`{}`),
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.createdApp == nil || repo.createdApp.DueAt == nil {
		t.Fatal("expected DueAt to be set")
	}
}

func TestSubmitApplication_Success_WithFiles(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 1024},
		&fakeMailer{},
	)

	files := []*multipart.FileHeader{{Filename: "id.pdf", Size: 1024}}
	resp, err := svc.SubmitApplication("u1", validReq(), files)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Attachments, 1)
	assert.Equal(t, "id.pdf", resp.Attachments[0].FileName)
}

func TestSubmitApplication_LogFailureDoesNotBreakMainFlow(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 100},
		&fakeMailer{},
		&fakeActivityLogger{err: errors.New("log failed")},
	)

	resp, err := svc.SubmitApplication("u1", validReq(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSubmitApplication_PublishesApplicationSubmittedEvent(t *testing.T) {
	repo := &fakeAppRepo{}
	pub := &fakeApplicationEventPublisher{}
	svc := newSvc(
		repo,
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{},
		&fakeMailer{},
	).WithEventPublisher(pub)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pub.published {
		t.Fatal("expected publish call")
	}
	if pub.lastSubmitted.CitizenUserID != "u1" {
		t.Fatalf("expected citizen u1, got %q", pub.lastSubmitted.CitizenUserID)
	}
}

func TestSubmitApplication_PublishFailureDoesNotFailRequest(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{},
		&fakeMailer{},
	).WithEventPublisher(&fakeApplicationEventPublisher{err: errors.New("publish failed")})

	resp, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSubmitApplication_ServiceTypeNotFound(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{err: errors.New("not found")},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.ErrorIs(t, err, ErrServiceTypeNotFound)
}

func TestSubmitApplication_NilRequest(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", nil, nil)
	assert.ErrorIs(t, err, ErrApplicationInvalidInput)
}

func TestSubmitApplication_ServiceTypeNil(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: nil, err: nil},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.ErrorIs(t, err, ErrServiceTypeNotFound)
}

func TestSubmitApplication_ServiceTypeInactive(t *testing.T) {
	st := activeServiceType()
	st.IsActive = false
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.ErrorIs(t, err, ErrServiceTypeInactive)
}

func TestSubmitApplication_MissingRequiredField(t *testing.T) {
	st := activeServiceType()
	st.FormSchema = makeSchema([]string{"full_name", "date_of_birth"})

	req := &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{"full_name": "An"}), // missing date_of_birth
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", req, nil)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestSubmitApplication_TooManyAttachments(t *testing.T) {
	files := make([]*multipart.FileHeader, maxAttachments+1)
	for i := range files {
		files[i] = &multipart.FileHeader{Size: 1024}
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrTooManyAttachments)
}

func TestSubmitApplication_FileTooLarge(t *testing.T) {
	files := []*multipart.FileHeader{
		{Size: maxFileSizeBytes + 1},
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentTooLarge)
}

func TestSubmitApplication_TotalSizeTooLarge(t *testing.T) {
	// 3 files × 11 MiB each = 33 MiB > 30 MiB limit
	const elevenMiB = 11 * 1024 * 1024
	files := []*multipart.FileHeader{
		{Size: elevenMiB},
		{Size: elevenMiB},
		{Size: elevenMiB},
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentTooLarge)
}

func TestSubmitApplication_InvalidMimeType(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "malware.exe", Size: 1024}}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: utils.ErrDisallowedMime},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentInvalidType)
}

func TestSubmitApplication_StorageError(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "doc.pdf", Size: 1024}}
	storageErr := errors.New("disk full")

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: storageErr},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrAttachmentInvalidType))
}

func TestSubmitApplication_RepoCreateError(t *testing.T) {
	repoErr := errors.New("db constraint violation")
	svc := newSvc(
		&fakeAppRepo{createErr: repoErr},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.Error(t, err)
}

func TestSubmitApplication_NoRequiredFieldsInSchema(t *testing.T) {
	st := activeServiceType()
	st.FormSchema = makeSchema(nil) // no required fields

	req := &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{}),
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com"}},
		&fakeStorage{},
		&fakeMailer{},
	)

	resp, err := svc.SubmitApplication("u1", req, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// --- ListMyApplications ---

func TestAppService_ListMyApplications_Success(t *testing.T) {
	apps := []models.Application{
		{ID: "app1", ApplicationCode: "APP-20240101-ABCDEF", SubmittedAt: time.Now()},
	}
	svc := newSvc(
		&fakeAppRepo{apps: apps, total: 1},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	result, total, err := svc.ListMyApplications("u1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAppService_ListMyApplications_Error(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newSvc(
		&fakeAppRepo{listErr: repoErr},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, _, err := svc.ListMyApplications("u1", 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetMyApplication ---

func TestAppService_GetMyApplication_Success(t *testing.T) {
	app := &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-20240101-ABCDEF",
		ServiceType:     models.ServiceType{Name: "Cấp CCCD"},
		Status:          models.ApplicationStatusReceived,
		SubmittedAt:     time.Now(),
	}
	svc := newSvc(
		&fakeAppRepo{app: app},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	resp, err := svc.GetMyApplication("u1", "app1")
	assert.NoError(t, err)
	assert.Equal(t, "app1", resp.ID)
	assert.Equal(t, "APP-20240101-ABCDEF", resp.ApplicationCode)
	assert.Equal(t, "Cấp CCCD", resp.ServiceTypeName)
}

func TestAppService_GetMyApplication_WithAttachments(t *testing.T) {
	sz := int64(1024)
	app := &models.Application{
		ID:              "app2",
		ApplicationCode: "APP-20240102-XYZ",
		ServiceType:     models.ServiceType{Name: "Cấp CCCD"},
		Status:          models.ApplicationStatusReceived,
		ApplicationAttachments: []models.ApplicationAttachment{
			{ID: "att-1", FileName: "id.pdf", FileURL: "/uploads/id.pdf", FileType: "application/pdf", FileSize: &sz},
		},
	}
	svc := newSvc(
		&fakeAppRepo{app: app},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	resp, err := svc.GetMyApplication("u1", "app2")
	assert.NoError(t, err)
	assert.Len(t, resp.Attachments, 1)
	assert.Equal(t, "id.pdf", resp.Attachments[0].FileName)
}

func TestAppService_GetMyApplication_NotFound(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{getErr: gorm.ErrRecordNotFound},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.GetMyApplication("u1", "missing")
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_GetMyApplication_InternalError(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{getErr: errors.New("db error")},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)
	_, err := svc.GetMyApplication("u1", "app-bad")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_ListMyApplicationStatusHistory_Success(t *testing.T) {
	logs := []models.ApplicationStatusLog{{ID: "log-1", NewStatus: models.ApplicationStatusProcessing}}
	svc := newSvc(
		&fakeAppRepo{
			app:       &models.Application{ID: "app-1", CitizenUserID: "u1"},
			logs:      logs,
			logsTotal: 1,
		},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	items, total, err := svc.ListMyApplicationStatusHistory("u1", "app-1", 1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
}

func TestAppService_ListMyApplicationStatusHistory_NotFound(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{getErr: errors.New("record not found")},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, _, err := svc.ListMyApplicationStatusHistory("u1", "bad", 1, 10, nil)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_ListMyApplicationStatusHistory_DBError(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{getErr: errors.New("db down")},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, _, err := svc.ListMyApplicationStatusHistory("u1", "bad", 1, 10, nil)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_ListMyApplicationStatusHistory_NilApp(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{app: nil, getErr: nil},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)
	_, _, err := svc.ListMyApplicationStatusHistory("u1", "app-x", 1, 10, nil)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_UploadMyApplicationSupplements_Success(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "bo-sung.pdf", Size: 1024}}
	svc := newSvc(
		&fakeAppRepo{app: &models.Application{ID: "app-1", Status: models.ApplicationStatusNeedMoreInfo}},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{pubURL: "/uploads/app-1/bo-sung.pdf", mime: "application/pdf", size: 1024},
		&fakeMailer{},
	)

	resp, err := svc.UploadMyApplicationSupplements("u1", "app-1", files)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "bo-sung.pdf", resp[0].FileName)
}

func TestAppService_UploadMyApplicationSupplements_EmptyFiles(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.UploadMyApplicationSupplements("u1", "app-1", nil)
	assert.ErrorIs(t, err, ErrAttachmentRequired)
}

func TestAppService_UploadMyApplicationSupplements_StatusNotAllowed(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "bo-sung.pdf", Size: 1024}}
	svc := newSvc(
		&fakeAppRepo{app: &models.Application{ID: "app-1", Status: models.ApplicationStatusApproved}},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.UploadMyApplicationSupplements("u1", "app-1", files)
	assert.ErrorIs(t, err, ErrSupplementNotAllowed)
}

func TestAppService_UploadMyApplicationSupplements_NotFound(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "bo-sung.pdf", Size: 1024}}
	svc := newSvc(
		&fakeAppRepo{getErr: errors.New("record not found")},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.UploadMyApplicationSupplements("u1", "missing", files)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_UploadMyApplicationSupplements_DBError(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "bo-sung.pdf", Size: 1024}}
	svc := newSvc(
		&fakeAppRepo{getErr: errors.New("db down")},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.UploadMyApplicationSupplements("u1", "missing", files)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_UploadMyApplicationSupplements_NilApp(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "f.pdf", Size: 100}}
	svc := newSvc(
		&fakeAppRepo{app: nil, getErr: nil},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)
	_, err := svc.UploadMyApplicationSupplements("u1", "app-x", files)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestAppService_UploadMyApplicationSupplements_SaveErrDisallowedMime(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "f.exe", Size: 100}}
	svc := newSvc(
		&fakeAppRepo{app: &models.Application{ID: "app-1", Status: models.ApplicationStatusNeedMoreInfo}},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: utils.ErrDisallowedMime},
		&fakeMailer{},
	)
	_, err := svc.UploadMyApplicationSupplements("u1", "app-1", files)
	assert.ErrorIs(t, err, ErrAttachmentInvalidType)
}

func TestAppService_UploadMyApplicationSupplements_SaveErrGeneric(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "f.pdf", Size: 100}}
	svc := newSvc(
		&fakeAppRepo{app: &models.Application{ID: "app-1", Status: models.ApplicationStatusNeedMoreInfo}},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: errors.New("disk full")},
		&fakeMailer{},
	)
	_, err := svc.UploadMyApplicationSupplements("u1", "app-1", files)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrAttachmentInvalidType)
}

func TestAppService_UploadMyApplicationSupplements_CreateAttachmentsError(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "f.pdf", Size: 100}}
	svc := newSvc(
		&fakeAppRepo{
			app:          &models.Application{ID: "app-1", Status: models.ApplicationStatusNeedMoreInfo},
			createAttErr: errors.New("db error"),
		},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 100},
		&fakeMailer{},
	)
	_, err := svc.UploadMyApplicationSupplements("u1", "app-1", files)
	assert.Error(t, err)
}

// --- validateSubmittedData ---

func TestValidateSubmittedData_EmptySchema(t *testing.T) {
	err := validateSubmittedData(makeData(nil), json.RawMessage{})
	assert.NoError(t, err)
}

func TestValidateSubmittedData_AllPresent(t *testing.T) {
	schema := makeSchema([]string{"name", "dob"})
	data := makeData(map[string]string{"name": "An", "dob": "1990-01-01"})
	assert.NoError(t, validateSubmittedData(data, schema))
}

func TestValidateSubmittedData_MissingField(t *testing.T) {
	schema := makeSchema([]string{"name", "dob"})
	data := makeData(map[string]string{"name": "An"})
	err := validateSubmittedData(data, schema)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestValidateSubmittedData_EmptyStringField(t *testing.T) {
	schema := makeSchema([]string{"name"})
	data := makeData(map[string]string{"name": ""})
	err := validateSubmittedData(data, schema)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestValidateSubmittedData_InvalidJSON(t *testing.T) {
	schema := makeSchema([]string{"name"})
	err := validateSubmittedData(json.RawMessage(`{bad json`), schema)
	assert.ErrorIs(t, err, ErrInvalidSubmittedData)
}
