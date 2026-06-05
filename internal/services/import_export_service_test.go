package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// --- fakes ---

type fakeIEDeptRepo struct {
	createErr error
}

func (r *fakeIEDeptRepo) FindByID(_ context.Context, _ string) (*models.Department, error) {
	return nil, nil
}
func (r *fakeIEDeptRepo) FindByCode(_ context.Context, _ string) (*models.Department, error) {
	return nil, nil
}
func (r *fakeIEDeptRepo) FindByLeaderUserID(_ context.Context, _ string) (*models.Department, error) {
	return nil, nil
}
func (r *fakeIEDeptRepo) Create(_ context.Context, d *models.Department) (*models.Department, error) {
	return d, nil
}
func (r *fakeIEDeptRepo) CreateInTx(_ *gorm.DB, _ *models.Department) error { return r.createErr }
func (r *fakeIEDeptRepo) Update(_ context.Context, _ *models.Department) error {
	return nil
}
func (r *fakeIEDeptRepo) List(_ context.Context, _ repositories.DepartmentFilter, _, _ int) ([]models.Department, int64, error) {
	return nil, 0, nil
}
func (r *fakeIEDeptRepo) SoftDelete(_ context.Context, _ string, _ string) error { return nil }

var _ repositories.DepartmentRepository = (*fakeIEDeptRepo)(nil)

type fakeIEUserRepo struct {
	createErr error
}

func (r *fakeIEUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeIEUserRepo) FindByID(_ string) (*models.User, error)     { return nil, nil }
func (r *fakeIEUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeIEUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return r.createErr }
func (r *fakeIEUserRepo) Update(_ *models.User) error                 { return nil }
func (r *fakeIEUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeIEUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeIEUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

var _ repositories.UserRepository = (*fakeIEUserRepo)(nil)

type fakeIEProfileRepo struct {
	createErr error
}

func (r *fakeIEProfileRepo) GetByUserID(_ string) (*models.CitizenProfile, error) { return nil, nil }
func (r *fakeIEProfileRepo) Update(_ *models.CitizenProfile) error                { return nil }
func (r *fakeIEProfileRepo) CreateInTx(_ *gorm.DB, _ *models.CitizenProfile) error {
	return r.createErr
}
func (r *fakeIEProfileRepo) FindByCitizenIDNumber(_ string) (*models.CitizenProfile, error) {
	return nil, nil
}
func (r *fakeIEProfileRepo) ListAllForExport(_, _ int) ([]repositories.CitizenExportRow, int64, error) {
	return nil, 0, nil
}

var _ repositories.CitizenProfileRepository = (*fakeIEProfileRepo)(nil)

type fakeIEServiceTypeRepo struct {
	createErr error
}

func (r *fakeIEServiceTypeRepo) List(_ context.Context, _ repositories.ListFilter) (*repositories.ListResult, error) {
	return nil, nil
}
func (r *fakeIEServiceTypeRepo) GetByID(_ context.Context, _ string) (*models.ServiceType, error) {
	return nil, nil
}
func (r *fakeIEServiceTypeRepo) GetByIDForAdmin(_ context.Context, _ string) (*models.ServiceType, error) {
	return nil, nil
}
func (r *fakeIEServiceTypeRepo) ListDepartments(_ context.Context) ([]models.Department, error) {
	return nil, nil
}
func (r *fakeIEServiceTypeRepo) ListCategories(_ context.Context) ([]models.Category, error) {
	return nil, nil
}
func (r *fakeIEServiceTypeRepo) Create(_ context.Context, _ *models.ServiceType) error { return nil }
func (r *fakeIEServiceTypeRepo) CreateInTx(_ *gorm.DB, _ *models.ServiceType) error {
	return r.createErr
}
func (r *fakeIEServiceTypeRepo) Update(_ context.Context, _ *models.ServiceType) error { return nil }
func (r *fakeIEServiceTypeRepo) Delete(_ context.Context, _ string) error              { return nil }
func (r *fakeIEServiceTypeRepo) CountApplications(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

var _ repositories.ServiceTypeRepository = (*fakeIEServiceTypeRepo)(nil)

type fakeIEStaffProfileRepo struct {
	createErr error
}

func (r *fakeIEStaffProfileRepo) FindByUserID(_ string) (*models.StaffProfile, error) {
	return nil, nil
}
func (r *fakeIEStaffProfileRepo) ListByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return nil, 0, nil
}
func (r *fakeIEStaffProfileRepo) UpdateDepartment(_ string, _ *string, _ string) error { return nil }
func (r *fakeIEStaffProfileRepo) Create(p *models.StaffProfile) (*models.StaffProfile, error) {
	return p, nil
}
func (r *fakeIEStaffProfileRepo) CreateInTx(_ *gorm.DB, _ *models.StaffProfile) error {
	return r.createErr
}

var _ repositories.StaffProfileRepository = (*fakeIEStaffProfileRepo)(nil)

type fakeIETransactor struct {
	err error
}

func (t *fakeIETransactor) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	if t.err != nil {
		return t.err
	}
	return fc(nil)
}

func newIESvc(deptRepo *fakeIEDeptRepo, userRepo *fakeIEUserRepo, profileRepo *fakeIEProfileRepo, stRepo *fakeIEServiceTypeRepo) *ImportExportService {
	return NewImportExportService(&fakeIETransactor{}, deptRepo, userRepo, profileRepo, stRepo, &fakeIEStaffProfileRepo{})
}

// --- ImportDepartments ---

func TestImportExportService_ImportDepartments_OK(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []DepartmentImportRow{
		{Ten: "Phòng IT", MaCode: "IT"},
		{Ten: "Phòng Hành chính", MaCode: "HC"},
	}
	errs := svc.ImportDepartments(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportDepartments_MissingName(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []DepartmentImportRow{{Ten: "", MaCode: "IT"}}
	errs := svc.ImportDepartments(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Tên")
}

func TestImportExportService_ImportDepartments_MissingCode(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []DepartmentImportRow{{Ten: "IT Dept", MaCode: ""}}
	errs := svc.ImportDepartments(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Mã code")
}

func TestImportExportService_ImportDepartments_DBError(t *testing.T) {
	deptRepo := &fakeIEDeptRepo{createErr: errors.New("duplicate key")}
	svc := newIESvc(deptRepo, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []DepartmentImportRow{{Ten: "IT", MaCode: "IT"}}
	errs := svc.ImportDepartments(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

// --- ImportStaff ---

func TestImportExportService_ImportStaff_OK(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{
		{HoTen: "Nguyễn Văn B", Email: "b@b.com", VaiTro: "staff"},
	}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportStaff_InvalidRole(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{
		{HoTen: "Test", Email: "t@t.com", VaiTro: "invalid_role"},
	}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Vai trò")
}

func TestImportExportService_ImportStaff_MissingName(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{{HoTen: "", Email: "t@t.com", VaiTro: "staff"}}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Họ tên")
}

func TestImportExportService_ImportStaff_DBError(t *testing.T) {
	userRepo := &fakeIEUserRepo{createErr: errors.New("duplicate email")}
	svc := newIESvc(&fakeIEDeptRepo{}, userRepo, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{{HoTen: "Test", Email: "t@t.com", VaiTro: "staff"}}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportStaff_MissingEmail(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{{HoTen: "Test", Email: "", VaiTro: "staff"}}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

// --- ImportCitizens ---

func TestImportExportService_ImportCitizens_OK(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{
		{SoCCCD: "123456789012", HoTen: "Nguyễn Văn A", Email: "a@a.com"},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportCitizens_MissingCCCD(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{{SoCCCD: "", HoTen: "Test", Email: "t@t.com"}}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Số CCCD")
}

func TestImportExportService_ImportCitizens_InvalidCCCD(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{
		{SoCCCD: "123", HoTen: "Test", Email: "t@t.com"},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "12 chữ số")
}

func TestImportExportService_ImportCitizens_MissingName(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{{SoCCCD: "123456789012", HoTen: "", Email: "a@a.com"}}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Họ tên")
}

func TestImportExportService_ImportCitizens_DBError(t *testing.T) {
	userRepo := &fakeIEUserRepo{createErr: errors.New("duplicate email")}
	svc := newIESvc(&fakeIEDeptRepo{}, userRepo, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{{SoCCCD: "123456789012", HoTen: "Test", Email: "a@a.com"}}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportCitizens_ProfileDBError(t *testing.T) {
	profileRepo := &fakeIEProfileRepo{createErr: errors.New("profile insert failed")}
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, profileRepo, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{{SoCCCD: "123456789012", HoTen: "Test", Email: "a@a.com"}}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportCitizens_MissingEmail(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{
		{SoCCCD: "123456789012", HoTen: "Test", Email: ""},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportCitizens_WithDateOfBirth(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{
		{SoCCCD: "123456789012", HoTen: "Nguyễn Văn A", Email: "a@a.com", NgaySinh: "15/06/1990"},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportCitizens_WithDateOfBirthISO(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []CitizenImportRow{
		{SoCCCD: "123456789012", HoTen: "Nguyễn Văn A", Email: "a@a.com", NgaySinh: "1990-06-15"},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportCitizens_DuplicateKeyError(t *testing.T) {
	svc := NewImportExportService(
		&fakeIETransactor{err: gorm.ErrDuplicatedKey},
		&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{}, &fakeIEStaffProfileRepo{},
	)
	rows := []CitizenImportRow{
		{SoCCCD: "123456789012", HoTen: "Nguyễn Văn A", Email: "a@a.com"},
	}
	errs := svc.ImportCitizens(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Dữ liệu bị trùng")
}

// --- ImportServiceTypes ---

func TestImportExportService_ImportServiceTypes_OK(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []ServiceTypeImportRow{
		{Ten: "Cấp hộ khẩu"},
	}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportServiceTypes_DBError(t *testing.T) {
	stRepo := &fakeIEServiceTypeRepo{createErr: errors.New("db error")}
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, stRepo)
	rows := []ServiceTypeImportRow{{Ten: "Cấp hộ khẩu"}}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportServiceTypes_LongName(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	longName := strings.Repeat("A", 110) // triggers utf8.RuneCountInString(code) > 100 branch
	rows := []ServiceTypeImportRow{{Ten: longName}}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportStaff_DeptNotFound(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []StaffImportRow{
		{HoTen: "Test", Email: "t@t.com", VaiTro: "staff", MaPhongBan: "PB999"},
	}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "không tìm thấy phòng ban")
}

func TestImportExportService_ImportStaff_DeptLookupError(t *testing.T) {
	deptRepo := &fakeIEDeptRepoWithFindByCodeErr{err: errors.New("db error")}
	svc := NewImportExportService(&fakeIETransactor{}, deptRepo, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{}, &fakeIEStaffProfileRepo{})
	rows := []StaffImportRow{
		{HoTen: "Test", Email: "t@t.com", VaiTro: "staff", MaPhongBan: "PB001"},
	}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "lỗi tra cứu phòng ban")
}

func TestImportExportService_ImportStaff_StaffProfileDBError(t *testing.T) {
	staffProfileRepo := &fakeIEStaffProfileRepo{createErr: errors.New("profile error")}
	svc := NewImportExportService(&fakeIETransactor{}, &fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{}, staffProfileRepo)
	rows := []StaffImportRow{{HoTen: "Test", Email: "t@t.com", VaiTro: "staff"}}
	errs := svc.ImportStaff(rows, "admin-1")
	assert.NotEmpty(t, errs)
}

func TestImportExportService_ImportServiceTypes_MissingName(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []ServiceTypeImportRow{{Ten: ""}}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "Tên")
}

func TestImportExportService_ImportServiceTypes_WithProcessingTimeAndFee(t *testing.T) {
	svc := newIESvc(&fakeIEDeptRepo{}, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []ServiceTypeImportRow{
		{Ten: "Dịch vụ", ThoiGianXuLyNgay: "5", Phi: "100000"},
	}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.Empty(t, errs)
}

func TestImportExportService_ImportServiceTypes_DeptNotFound(t *testing.T) {
	deptRepo := &fakeIEDeptRepo{}
	// FindByCode returns nil, nil by default (dept not found)
	svc := newIESvc(deptRepo, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{})
	rows := []ServiceTypeImportRow{
		{Ten: "Dịch vụ", MaPhongBan: "PB999"},
	}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "không tìm thấy phòng ban")
}

func TestImportExportService_ImportServiceTypes_DeptLookupError(t *testing.T) {
	deptRepo := &fakeIEDeptRepoWithFindByCodeErr{err: errors.New("db error")}
	svc := NewImportExportService(&fakeIETransactor{}, deptRepo, &fakeIEUserRepo{}, &fakeIEProfileRepo{}, &fakeIEServiceTypeRepo{}, &fakeIEStaffProfileRepo{})
	rows := []ServiceTypeImportRow{
		{Ten: "Dịch vụ", MaPhongBan: "PB001"},
	}
	errs := svc.ImportServiceTypes(rows, "admin-1")
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "lỗi tra cứu phòng ban")
}

type fakeIEDeptRepoWithFindByCodeErr struct {
	fakeIEDeptRepo
	err error
}

func (r *fakeIEDeptRepoWithFindByCodeErr) FindByCode(_ context.Context, _ string) (*models.Department, error) {
	return nil, r.err
}
