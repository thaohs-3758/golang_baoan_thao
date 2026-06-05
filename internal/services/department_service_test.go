package services

import (
	"context"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// --- fakeDepartmentRepo ---

type fakeDepartmentRepo struct {
	dept       *models.Department
	depts      []models.Department
	total      int64
	findErr    error
	codeErr    error
	codeDept   *models.Department
	leaderDept *models.Department
	createErr  error
	updateErr  error
	deleteErr  error
}

func (r *fakeDepartmentRepo) FindByID(_ context.Context, _ string) (*models.Department, error) {
	return r.dept, r.findErr
}
func (r *fakeDepartmentRepo) FindByCode(_ context.Context, _ string) (*models.Department, error) {
	return r.codeDept, r.codeErr
}
func (r *fakeDepartmentRepo) FindByLeaderUserID(_ context.Context, _ string) (*models.Department, error) {
	return r.leaderDept, r.findErr
}
func (r *fakeDepartmentRepo) Create(_ context.Context, d *models.Department) (*models.Department, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	if d.ID == "" {
		d.ID = "dept-new"
	}
	return d, nil
}
func (r *fakeDepartmentRepo) Update(_ context.Context, _ *models.Department) error {
	return r.updateErr
}
func (r *fakeDepartmentRepo) List(_ context.Context, _ repositories.DepartmentFilter, _, _ int) ([]models.Department, int64, error) {
	if r.findErr != nil {
		return nil, 0, r.findErr
	}
	return r.depts, r.total, nil
}
func (r *fakeDepartmentRepo) SoftDelete(_ context.Context, _ string, _ string) error {
	return r.deleteErr
}
func (r *fakeDepartmentRepo) CreateInTx(_ *gorm.DB, d *models.Department) error {
	if r.createErr != nil {
		return r.createErr
	}
	return nil
}

var _ repositories.DepartmentRepository = (*fakeDepartmentRepo)(nil)

type fakeStaffRepo struct {
	updateCalls []struct {
		userID string
		deptID *string
	}
	updateErr error
}

func (r *fakeStaffRepo) FindByUserID(_ string) (*models.StaffProfile, error) { return nil, nil }
func (r *fakeStaffRepo) ListByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return nil, 0, nil
}
func (r *fakeStaffRepo) UpdateDepartment(userID string, deptID *string, _ string) error {
	r.updateCalls = append(r.updateCalls, struct {
		userID string
		deptID *string
	}{userID: userID, deptID: deptID})
	return r.updateErr
}
func (r *fakeStaffRepo) Create(p *models.StaffProfile) (*models.StaffProfile, error) { return p, nil }
func (r *fakeStaffRepo) CreateInTx(_ *gorm.DB, _ *models.StaffProfile) error         { return nil }

var _ repositories.StaffProfileRepository = (*fakeStaffRepo)(nil)

func newDeptSvc(repo *fakeDepartmentRepo) *DepartmentService {
	return NewDepartmentService(repo, nil)
}

func newDeptSvcWithStaff(repo *fakeDepartmentRepo) *DepartmentService {
	return NewDepartmentService(repo, &fakeStaffRepo{})
}

// --- ListDepartments ---

func TestDepartmentService_ListDepartments_OK(t *testing.T) {
	depts := []models.Department{{ID: "1", Name: "IT"}}
	svc := newDeptSvc(&fakeDepartmentRepo{depts: depts, total: 1})
	result, total, err := svc.ListDepartments(repositories.DepartmentFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestDepartmentService_ListDepartments_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	_, _, err := svc.ListDepartments(repositories.DepartmentFilter{}, 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetDepartment ---

func TestDepartmentService_GetDepartment_Found(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "IT"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	result, err := svc.GetDepartment("d1")
	assert.NoError(t, err)
	assert.Equal(t, "d1", result.ID)
}

func TestDepartmentService_GetDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	_, err := svc.GetDepartment("missing")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_GetDepartment_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	_, err := svc.GetDepartment("x")
	assert.ErrorIs(t, err, repoErr)
}

// --- CreateDepartment ---

func TestDepartmentService_CreateDepartment_OK(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: nil})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", Address: "123 St"}
	dept, err := svc.CreateDepartment(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "IT", dept.Name)
	assert.Equal(t, "IT001", dept.Code)
}

func TestDepartmentService_CreateDepartment_WithLeader(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: nil})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", LeaderUserID: "user-1"}
	dept, err := svc.CreateDepartment(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", *dept.LeaderUserID)
}

func TestDepartmentService_CreateDepartment_WithLeaderAndStaffRepo(t *testing.T) {
	staffRepo := &fakeStaffRepo{}
	svc := NewDepartmentService(&fakeDepartmentRepo{codeDept: nil}, staffRepo)
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", LeaderUserID: "user-1"}
	dept, err := svc.CreateDepartment(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", *dept.LeaderUserID)
	if assert.Len(t, staffRepo.updateCalls, 1) {
		assert.Equal(t, "user-1", staffRepo.updateCalls[0].userID)
		if assert.NotNil(t, staffRepo.updateCalls[0].deptID) {
			assert.Equal(t, "dept-new", *staffRepo.updateCalls[0].deptID)
		}
	}
}

func TestDepartmentService_CreateDepartment_CodeExists(t *testing.T) {
	existing := &models.Department{Code: "IT001"}
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: existing})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentCodeExists)
}

func TestDepartmentService_CreateDepartment_LeaderAlreadyAssigned(t *testing.T) {
	existingLeaderDept := &models.Department{ID: "d-existing", LeaderUserID: ptrStrDept("user-1")}
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: nil, leaderDept: existingLeaderDept})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", LeaderUserID: "user-1"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentLeaderAlreadyAssigned)
}

func TestDepartmentService_CreateDepartment_FindCodeError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{codeErr: repoErr})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_CreateDepartment_CreateError(t *testing.T) {
	createErr := errors.New("create failed")
	svc := newDeptSvc(&fakeDepartmentRepo{createErr: createErr})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, createErr)
}

// --- UpdateDepartment ---

func TestDepartmentService_UpdateDepartment_OK(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "Old", Code: "OLD"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	req := &dtos.DepartmentUpdateRequest{Name: "New", Code: "OLD", Address: "New St"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
}

func TestDepartmentService_UpdateDepartment_CodeChange_Unique(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, codeDept: nil})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "NEW"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "NEW", result.Code)
}

func TestDepartmentService_UpdateDepartment_CodeChange_FindCodeError(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	codeErr := errors.New("code lookup failed")
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, codeErr: codeErr})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "NEW"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, codeErr)
}

func TestDepartmentService_UpdateDepartment_CodeChange_Conflict(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	existing := &models.Department{ID: "d2", Code: "NEW"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, codeDept: existing})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "NEW"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentCodeExists)
}

func TestDepartmentService_UpdateDepartment_LeaderAlreadyAssigned(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	existingLeaderDept := &models.Department{ID: "d2", LeaderUserID: ptrStrDept("user-1")}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, leaderDept: existingLeaderDept})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "OLD", LeaderUserID: "user-1"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentLeaderAlreadyAssigned)
}

func TestDepartmentService_UpdateDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "X"}
	_, err := svc.UpdateDepartment("missing", req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_UpdateDepartment_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "X"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_UpdateDepartment_SaveError(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	saveErr := errors.New("save failed")
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, updateErr: saveErr})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "OLD"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, saveErr)
}

func TestDepartmentService_UpdateDepartment_WithLeaderAndStaffRepo(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	staffRepo := &fakeStaffRepo{}
	svc := NewDepartmentService(&fakeDepartmentRepo{dept: d}, staffRepo)
	req := &dtos.DepartmentUpdateRequest{Name: "IT", Code: "OLD", LeaderUserID: "user-1"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", *result.LeaderUserID)
	if assert.Len(t, staffRepo.updateCalls, 1) {
		assert.Equal(t, "user-1", staffRepo.updateCalls[0].userID)
		if assert.NotNil(t, staffRepo.updateCalls[0].deptID) {
			assert.Equal(t, "d1", *staffRepo.updateCalls[0].deptID)
		}
	}
}

func TestDepartmentService_UpdateDepartment_ClearLeader(t *testing.T) {
	leaderID := "user-1"
	d := &models.Department{ID: "d1", Code: "OLD", LeaderUserID: &leaderID}
	staffRepo := &fakeStaffRepo{}
	svc := NewDepartmentService(&fakeDepartmentRepo{dept: d}, staffRepo)
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "OLD", LeaderUserID: ""}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Nil(t, result.LeaderUserID)
	if assert.Len(t, staffRepo.updateCalls, 1) {
		assert.Equal(t, "user-1", staffRepo.updateCalls[0].userID)
		assert.Nil(t, staffRepo.updateCalls[0].deptID)
	}
}

func TestDepartmentService_UpdateDepartment_ChangeLeader_ClearsOldAssignsNew(t *testing.T) {
	oldLeader := "user-old"
	d := &models.Department{ID: "d1", Code: "OLD", LeaderUserID: &oldLeader}
	staffRepo := &fakeStaffRepo{}
	svc := NewDepartmentService(&fakeDepartmentRepo{dept: d}, staffRepo)
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "OLD", LeaderUserID: "user-new"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "user-new", *result.LeaderUserID)
	if assert.Len(t, staffRepo.updateCalls, 2) {
		assert.Equal(t, "user-old", staffRepo.updateCalls[0].userID)
		assert.Nil(t, staffRepo.updateCalls[0].deptID)
		assert.Equal(t, "user-new", staffRepo.updateCalls[1].userID)
		if assert.NotNil(t, staffRepo.updateCalls[1].deptID) {
			assert.Equal(t, "d1", *staffRepo.updateCalls[1].deptID)
		}
	}
}

// --- DeleteDepartment ---

func TestDepartmentService_DeleteDepartment_OK(t *testing.T) {
	d := &models.Department{ID: "d1"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	err := svc.DeleteDepartment("d1", "actor")
	assert.NoError(t, err)
}

func TestDepartmentService_DeleteDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	err := svc.DeleteDepartment("missing", "actor")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_DeleteDepartment_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	err := svc.DeleteDepartment("d1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_DeleteDepartment_DeleteError(t *testing.T) {
	d := &models.Department{ID: "d1"}
	deleteErr := errors.New("delete error")
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, deleteErr: deleteErr})
	err := svc.DeleteDepartment("d1", "actor")
	assert.ErrorIs(t, err, deleteErr)
}

func ptrStrDept(v string) *string { return &v }

func TestDepartmentService_LogActivity_WithLogger(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "Test"}
	repo := &fakeDepartmentRepo{codeDept: nil, dept: d}
	logger := &fakeActivityLogger{}
	svc := NewDepartmentService(repo, nil, logger)

	req := &dtos.DepartmentCreateRequest{Name: "Test", Code: "TEST"}
	_, err := svc.CreateDepartment(req, "actor1")
	assert.NoError(t, err)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "department.create", logger.entries[0].Action)
	}
}

func TestDepartmentService_LogActivity_LoggerError(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "Test"}
	repo := &fakeDepartmentRepo{codeDept: nil, dept: d}
	logger := &fakeActivityLogger{err: assert.AnError}
	svc := NewDepartmentService(repo, nil, logger)

	req := &dtos.DepartmentCreateRequest{Name: "Test", Code: "TEST"}
	// Should succeed even if logger fails
	_, err := svc.CreateDepartment(req, "actor1")
	assert.NoError(t, err)
}
