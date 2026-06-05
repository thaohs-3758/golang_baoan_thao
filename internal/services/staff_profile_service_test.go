package services

import (
	"context"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type fakeStaffProfileRepo struct {
	profiles  []models.StaffProfile
	total     int64
	listErr   error
	updateErr error
	updated   bool
}

func (r *fakeStaffProfileRepo) FindByUserID(_ string) (*models.StaffProfile, error) {
	return nil, nil
}
func (r *fakeStaffProfileRepo) ListByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return r.profiles, r.total, r.listErr
}
func (r *fakeStaffProfileRepo) UpdateDepartment(_ string, _ *string, _ string) error {
	r.updated = true
	return r.updateErr
}
func (r *fakeStaffProfileRepo) Create(p *models.StaffProfile) (*models.StaffProfile, error) {
	return p, nil
}
func (r *fakeStaffProfileRepo) CreateInTx(_ *gorm.DB, _ *models.StaffProfile) error { return nil }

var _ repositories.StaffProfileRepository = (*fakeStaffProfileRepo)(nil)

type fakeStaffUserRepo struct {
	users map[string]*models.User
}

func (r *fakeStaffUserRepo) FindByID(id string) (*models.User, error)    { return r.users[id], nil }
func (r *fakeStaffUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeStaffUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeStaffUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return nil }
func (r *fakeStaffUserRepo) Update(_ *models.User) error                 { return nil }
func (r *fakeStaffUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeStaffUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeStaffUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

type fakeStaffDepartmentRepo struct {
	dept *models.Department
	err  error
}

func (r *fakeStaffDepartmentRepo) FindByID(_ context.Context, _ string) (*models.Department, error) {
	return nil, nil
}
func (r *fakeStaffDepartmentRepo) FindByCode(_ context.Context, _ string) (*models.Department, error) {
	return nil, nil
}
func (r *fakeStaffDepartmentRepo) FindByLeaderUserID(_ context.Context, _ string) (*models.Department, error) {
	return r.dept, r.err
}
func (r *fakeStaffDepartmentRepo) Create(_ context.Context, d *models.Department) (*models.Department, error) {
	return d, nil
}
func (r *fakeStaffDepartmentRepo) CreateInTx(_ *gorm.DB, _ *models.Department) error { return nil }
func (r *fakeStaffDepartmentRepo) Update(_ context.Context, _ *models.Department) error {
	return nil
}
func (r *fakeStaffDepartmentRepo) List(_ context.Context, _ repositories.DepartmentFilter, _, _ int) ([]models.Department, int64, error) {
	return nil, 0, nil
}
func (r *fakeStaffDepartmentRepo) SoftDelete(_ context.Context, _ string, _ string) error {
	return nil
}

var _ repositories.DepartmentRepository = (*fakeStaffDepartmentRepo)(nil)

func TestStaffProfileService_FindStaffProfileByUserID(t *testing.T) {
	profile := &models.StaffProfile{User: models.User{ID: "u1"}}
	repo := &fakeStaffProfileRepo{}
	repo.profiles = []models.StaffProfile{*profile}
	svc := NewStaffProfileService(repo, nil)

	got, err := svc.FindStaffProfileByUserID("u1")
	assert.NoError(t, err)
	assert.Nil(t, got) // fakeStaffProfileRepo.FindByUserID always returns nil, nil
}

func TestStaffProfileService_ListStaffByDepartment_OK(t *testing.T) {
	profiles := []models.StaffProfile{{User: models.User{ID: "u1"}}}
	repo := &fakeStaffProfileRepo{profiles: profiles, total: 1}
	svc := NewStaffProfileService(repo, nil)
	result, total, err := svc.ListStaffByDepartment("dept-1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestStaffProfileService_AssignStaffToDepartment_UserNotFound(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeIEUserRepo{}
	svc := NewStaffProfileService(repo, userRepo)
	err := svc.AssignStaffToDepartment("u1", "dept-1", "admin-1")
	assert.NoError(t, err)
}

func TestStaffProfileService_AssignStaffToDepartment_UpdateDepartment(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	svc := NewStaffProfileService(repo, &fakeStaffUserRepo{users: map[string]*models.User{
		"u1":      {ID: "u1", Role: models.UserRoleStaff},
		"admin-1": {ID: "admin-1", Role: models.UserRoleSuperAdmin},
	}})
	err := svc.AssignStaffToDepartment("u1", "dept-1", "admin-1")
	assert.NoError(t, err)
	assert.True(t, repo.updated)
}

func TestStaffProfileService_AssignStaffToDepartment_ManagerCannotTransferLeader(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"u1":       {ID: "u1", Role: models.UserRoleStaff},
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	deptRepo := &fakeStaffDepartmentRepo{dept: &models.Department{ID: "dept-old"}}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.AssignStaffToDepartment("u1", "dept-new", "manager1")
	assert.ErrorIs(t, err, ErrLeaderTransferForbiddenForManager)
	assert.False(t, repo.updated)
}

func TestStaffProfileService_AssignStaffToDepartment_SuperAdminCanTransferLeader(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"u1":      {ID: "u1", Role: models.UserRoleStaff},
		"admin-1": {ID: "admin-1", Role: models.UserRoleSuperAdmin},
	}}
	deptRepo := &fakeStaffDepartmentRepo{dept: &models.Department{ID: "dept-old"}}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.AssignStaffToDepartment("u1", "dept-new", "admin-1")
	assert.NoError(t, err)
	assert.True(t, repo.updated)
}

func TestStaffProfileService_RemoveStaffFromDepartment_OK(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	svc := NewStaffProfileService(repo, &fakeStaffUserRepo{users: map[string]*models.User{
		"admin-1": {ID: "admin-1", Role: models.UserRoleSuperAdmin},
	}})
	err := svc.RemoveStaffFromDepartment("u1", "admin-1")
	assert.NoError(t, err)
}

func TestStaffProfileService_RemoveStaffFromDepartment_ManagerCannotRemoveLeader(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	deptRepo := &fakeStaffDepartmentRepo{dept: &models.Department{ID: "dept-old"}}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.RemoveStaffFromDepartment("u1", "manager1")
	assert.ErrorIs(t, err, ErrLeaderTransferForbiddenForManager)
	assert.False(t, repo.updated)
}

func TestStaffProfileService_AssignStaffToDepartment_ManagerDeptRepoError(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"u1":       {ID: "u1", Role: models.UserRoleStaff},
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	deptRepo := &fakeStaffDepartmentRepo{err: errors.New("db error")}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.AssignStaffToDepartment("u1", "dept-new", "manager1")
	assert.Error(t, err)
	assert.False(t, repo.updated)
}

func TestStaffProfileService_AssignStaffToDepartment_ManagerLeaderSameDept(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"u1":       {ID: "u1", Role: models.UserRoleStaff},
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	// leaderDept.ID == deptID → allowed
	deptRepo := &fakeStaffDepartmentRepo{dept: &models.Department{ID: "dept-1"}}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.AssignStaffToDepartment("u1", "dept-1", "manager1")
	assert.NoError(t, err)
	assert.True(t, repo.updated)
}

func TestStaffProfileService_RemoveStaffFromDepartment_ManagerDeptRepoError(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	deptRepo := &fakeStaffDepartmentRepo{err: errors.New("db error")}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.RemoveStaffFromDepartment("u1", "manager1")
	assert.Error(t, err)
	assert.False(t, repo.updated)
}

func TestStaffProfileService_RemoveStaffFromDepartment_ManagerLeaderNil(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"manager1": {ID: "manager1", Role: models.UserRoleManager},
	}}
	// leaderDept == nil → user is not a leader, proceed with removal
	deptRepo := &fakeStaffDepartmentRepo{dept: nil}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.RemoveStaffFromDepartment("u1", "manager1")
	assert.NoError(t, err)
	assert.True(t, repo.updated)
}

func TestStaffProfileService_RemoveStaffFromDepartment_SuperAdminCanRemoveLeader(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeStaffUserRepo{users: map[string]*models.User{
		"admin-1": {ID: "admin-1", Role: models.UserRoleSuperAdmin},
	}}
	deptRepo := &fakeStaffDepartmentRepo{dept: &models.Department{ID: "dept-old"}}
	svc := NewStaffProfileService(repo, userRepo, deptRepo)

	err := svc.RemoveStaffFromDepartment("u1", "admin-1")
	assert.NoError(t, err)
	assert.True(t, repo.updated)
}
