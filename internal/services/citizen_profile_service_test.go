package services

import (
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- fakes ---

type fakeProfileUserRepo struct {
	user      *models.User
	err       error
	updateErr error
	updated   *models.User
}

func (r *fakeProfileUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeProfileUserRepo) FindByID(_ string) (*models.User, error)     { return r.user, r.err }
func (r *fakeProfileUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeProfileUserRepo) CreateInTx(_ *gorm.DB, u *models.User) error { return nil }
func (r *fakeProfileUserRepo) Update(u *models.User) error                 { r.updated = u; return r.updateErr }
func (r *fakeProfileUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeProfileUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeProfileUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

type fakeProfileCitizenProfileRepo struct {
	profile   *models.CitizenProfile
	getErr    error
	updateErr error
	updated   *models.CitizenProfile
}

func (r *fakeProfileCitizenProfileRepo) GetByUserID(_ string) (*models.CitizenProfile, error) {
	return r.profile, r.getErr
}
func (r *fakeProfileCitizenProfileRepo) Update(p *models.CitizenProfile) error {
	r.updated = p
	return r.updateErr
}
func (r *fakeProfileCitizenProfileRepo) CreateInTx(_ *gorm.DB, _ *models.CitizenProfile) error {
	return nil
}
func (r *fakeProfileCitizenProfileRepo) FindByCitizenIDNumber(_ string) (*models.CitizenProfile, error) {
	return nil, nil
}
func (r *fakeProfileCitizenProfileRepo) ListAllForExport(_, _ int) ([]repositories.CitizenExportRow, int64, error) {
	return nil, 0, nil
}

type fakeProfileApplicationRepo struct {
	apps  []models.Application
	total int64
	err   error
}

func (r *fakeProfileApplicationRepo) ListByCitizen(_ string, _, _ int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.err
}
func (r *fakeProfileApplicationRepo) GetByIDForCitizen(_, _ string) (*models.Application, error) {
	return nil, nil
}
func (r *fakeProfileApplicationRepo) ListStatusLogsByCitizen(_, _ string, _, _ int, _ *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (r *fakeProfileApplicationRepo) CreateAttachments(_ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeProfileApplicationRepo) CreateWithAttachments(_ *models.Application, _ []models.ApplicationAttachment, _ func() string) error {
	return nil
}
func (r *fakeProfileApplicationRepo) AdminList(_ repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.err
}
func (r *fakeProfileApplicationRepo) GetByID(id string) (*models.Application, error) {
	for i := range r.apps {
		if r.apps[i].ID == id {
			return &r.apps[i], r.err
		}
	}
	return nil, r.err
}
func (r *fakeProfileApplicationRepo) ListDueWithin(_, _ time.Time) ([]models.Application, error) {
	return r.apps, r.err
}
func (r *fakeProfileApplicationRepo) AttachmentExistsForApplication(_ string) (bool, error) {
	return false, nil
}
func (r *fakeProfileApplicationRepo) UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error {
	return nil
}
func (r *fakeProfileApplicationRepo) ProcessStatusUpdate(_ string, _ *models.ApplicationStatus, _ models.ApplicationStatus, _ string, _ string, _, _ *time.Time, _ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeProfileApplicationRepo) GetDashboardStats() (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeProfileApplicationRepo) ListRecent(_ int) ([]models.Application, error) {
	return nil, nil
}
func (r *fakeProfileApplicationRepo) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeProfileApplicationRepo) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return nil, nil
}

// compile-time interface checks
var _ repositories.UserRepository = (*fakeProfileUserRepo)(nil)
var _ repositories.CitizenProfileRepository = (*fakeProfileCitizenProfileRepo)(nil)
var _ repositories.ApplicationRepository = (*fakeProfileApplicationRepo)(nil)

// --- GetProfile tests ---

func TestGetProfile_Success(t *testing.T) {
	dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com", Phone: "0912345678"}
	profile := &models.CitizenProfile{
		UserID:          "u1",
		CitizenIDNumber: "123456789012",
		DateOfBirth:     &dob,
		Gender:          "Nam",
	}

	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{profile: profile},
		&fakeProfileApplicationRepo{},
	)
	resp, err := svc.GetProfile("u1")

	assert.NoError(t, err)
	assert.Equal(t, "An", resp.Name)
	assert.Equal(t, "123456789012", resp.CitizenIDNumber)
	assert.Equal(t, "Nam", resp.Gender)
}

func TestGetProfile_UserNotFound(t *testing.T) {
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: nil},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.GetProfile("missing")
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestGetProfile_NoProfile_ReturnsEmptyFields(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{profile: nil},
		&fakeProfileApplicationRepo{},
	)
	resp, err := svc.GetProfile("u1")
	assert.NoError(t, err)
	assert.Empty(t, resp.CitizenIDNumber)
	assert.Equal(t, "An", resp.Name)
}

func TestGetProfile_UserRepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{err: repoErr},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.GetProfile("u1")
	assert.ErrorIs(t, err, repoErr)
}

func TestGetProfile_ProfileRepoError(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	repoErr := errors.New("db error")
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{getErr: repoErr},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.GetProfile("u1")
	assert.ErrorIs(t, err, repoErr)
}

// --- UpdateProfile tests ---

func TestUpdateProfile_UpdateNameAndPhone(t *testing.T) {
	user := &models.User{ID: "u1", Name: "Old Name", Email: "an@example.com"}
	profile := &models.CitizenProfile{UserID: "u1", CitizenIDNumber: "123456789012"}

	userRepo := &fakeProfileUserRepo{user: user}
	profileRepo := &fakeProfileCitizenProfileRepo{profile: profile}
	svc := NewCitizenProfileService(userRepo, profileRepo, &fakeProfileApplicationRepo{})

	newName := "New Name"
	newPhone := "0999888777"
	resp, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{
		Name:  &newName,
		Phone: &newPhone,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "0999888777", resp.Phone)
	assert.Equal(t, "New Name", userRepo.updated.Name)
	assert.Equal(t, "0999888777", userRepo.updated.Phone)
}

func TestUpdateProfile_AllOptionalFields(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	profile := &models.CitizenProfile{UserID: "u1", CitizenIDNumber: "123456789012"}
	profileRepo := &fakeProfileCitizenProfileRepo{profile: profile}
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		profileRepo,
		&fakeProfileApplicationRepo{},
	)

	addr := "123 Street"
	gender := "Nam"
	permanent := "456 Town"
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	enabled := false
	resp, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{
		Address:                  &addr,
		Gender:                   &gender,
		PermanentAddress:         &permanent,
		DateOfBirth:              &dob,
		EmailNotificationEnabled: &enabled,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Nam", resp.Gender)
	assert.Equal(t, "456 Town", resp.PermanentAddress)
	assert.Equal(t, false, resp.EmailNotificationEnabled)
}

func TestUpdateProfile_ProfileNotFound(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{profile: nil},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{})
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: nil, err: nil},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{})
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestUpdateProfile_UserRepoFindByIDError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{err: repoErr},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{})
	assert.ErrorIs(t, err, repoErr)
}

func TestUpdateProfile_ProfileRepoGetByUserIDError(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	repoErr := errors.New("db error")
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{getErr: repoErr},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{})
	assert.ErrorIs(t, err, repoErr)
}

func TestUpdateProfile_UserRepoUpdateError(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	profile := &models.CitizenProfile{UserID: "u1", CitizenIDNumber: "123456789012"}
	updateErr := errors.New("db error")
	newName := "New"
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user, updateErr: updateErr},
		&fakeProfileCitizenProfileRepo{profile: profile},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{Name: &newName})
	assert.ErrorIs(t, err, updateErr)
}

func TestUpdateProfile_ProfileRepoUpdateError(t *testing.T) {
	user := &models.User{ID: "u1", Name: "An", Email: "an@example.com"}
	profile := &models.CitizenProfile{UserID: "u1", CitizenIDNumber: "123456789012"}
	updateErr := errors.New("db error")
	newName := "New"
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{profile: profile, updateErr: updateErr},
		&fakeProfileApplicationRepo{},
	)
	_, err := svc.UpdateProfile("u1", &dtos.UpdateCitizenProfileRequest{Name: &newName})
	assert.ErrorIs(t, err, updateErr)
}

// --- ListMyApplications tests ---

func TestListMyApplications_Success(t *testing.T) {
	user := &models.User{ID: "u1"}
	apps := []models.Application{{ID: "app1", ApplicationCode: "APP-20240101-ABCDEF"}}
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{apps: apps, total: 1},
	)
	result, total, err := svc.ListMyApplications("u1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestListMyApplications_Error(t *testing.T) {
	repoErr := errors.New("db error")
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{err: repoErr},
	)
	_, _, err := svc.ListMyApplications("u1", 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

func TestChangeMyPassword_ConfirmationMismatch(t *testing.T) {
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: &models.User{ID: "u1"}},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "oldpass123",
		NewPassword:        "newpass123",
		ConfirmNewPassword: "differentpass",
	}
	err := svc.ChangeMyPassword("u1", req)
	assert.ErrorIs(t, err, ErrPasswordConfirmationMismatch)
}

func TestChangeMyPassword_NewEqualsCurrent(t *testing.T) {
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: &models.User{ID: "u1"}},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "oldpass123",
		NewPassword:        "oldpass123",
		ConfirmNewPassword: "oldpass123",
	}
	err := svc.ChangeMyPassword("u1", req)
	assert.ErrorIs(t, err, ErrNewPasswordMustDiffer)
}

func TestChangeMyPassword_CurrentMismatch(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("realoldpass"), bcrypt.DefaultCost)
	user := &models.User{ID: "u1", PasswordHash: string(hash)}
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: user},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "wrongoldpass",
		NewPassword:        "newpass123",
		ConfirmNewPassword: "newpass123",
	}
	err := svc.ChangeMyPassword("u1", req)
	assert.ErrorIs(t, err, ErrPasswordMismatch)
}

func TestChangeMyPassword_UserNotFound(t *testing.T) {
	svc := NewCitizenProfileService(
		&fakeProfileUserRepo{user: nil},
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "oldpass123",
		NewPassword:        "newpass123",
		ConfirmNewPassword: "newpass123",
	}
	err := svc.ChangeMyPassword("u1", req)
	assert.ErrorIs(t, err, ErrProfileNotFound)
}

func TestChangeMyPassword_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass123"), bcrypt.DefaultCost)
	user := &models.User{ID: "u1", PasswordHash: string(hash)}
	userRepo := &fakeProfileUserRepo{user: user}
	svc := NewCitizenProfileService(
		userRepo,
		&fakeProfileCitizenProfileRepo{},
		&fakeProfileApplicationRepo{},
	)
	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "oldpass123",
		NewPassword:        "newpass123",
		ConfirmNewPassword: "newpass123",
	}
	err := svc.ChangeMyPassword("u1", req)
	assert.NoError(t, err)
	assert.NotNil(t, userRepo.updated)
	err = bcrypt.CompareHashAndPassword([]byte(userRepo.updated.PasswordHash), []byte("newpass123"))
	assert.NoError(t, err)
}
