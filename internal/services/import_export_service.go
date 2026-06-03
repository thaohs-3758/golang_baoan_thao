package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var cccdRegex = regexp.MustCompile(`^\d{12}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// --- row types ---

type DepartmentImportRow struct {
	Ten    string
	MoTa   string
	MaCode string
}

type StaffImportRow struct {
	HoTen       string
	Email       string
	SoCCCD      string
	SoDienThoai string
	VaiTro      string
	MaPhongBan  string
}

type ServiceTypeImportRow struct {
	Ten              string
	MoTa             string
	ThoiGianXuLyNgay string
	Phi              string
	MaPhongBan       string
}

type CitizenImportRow struct {
	SoCCCD      string
	HoTen       string
	Email       string
	SoDienThoai string
	DiaChi      string
	NgaySinh    string
}

// --- import/export service ---

type ImportExportService struct {
	db               Transactor
	deptRepo         repositories.DepartmentRepository
	userRepo         repositories.UserRepository
	profileRepo      repositories.CitizenProfileRepository
	serviceRepo      repositories.ServiceTypeRepository
	staffProfileRepo repositories.StaffProfileRepository
}

func NewImportExportService(
	db Transactor,
	deptRepo repositories.DepartmentRepository,
	userRepo repositories.UserRepository,
	profileRepo repositories.CitizenProfileRepository,
	serviceRepo repositories.ServiceTypeRepository,
	staffProfileRepo repositories.StaffProfileRepository,
) *ImportExportService {
	return &ImportExportService{
		db:               db,
		deptRepo:         deptRepo,
		userRepo:         userRepo,
		profileRepo:      profileRepo,
		serviceRepo:      serviceRepo,
		staffProfileRepo: staffProfileRepo,
	}
}

// --- validation helpers ---

func validateRequired(val, field string, row int) string {
	if strings.TrimSpace(val) == "" {
		return fmt.Sprintf("Dòng %d: %s không được để trống", row, field)
	}
	return ""
}

// --- Department import ---

func (s *ImportExportService) ImportDepartments(rows []DepartmentImportRow, createdBy string) []string {
	var errs []string
	for i, r := range rows {
		lineNo := i + 2 // header is row 1
		if e := validateRequired(r.Ten, "Tên", lineNo); e != "" {
			errs = append(errs, e)
		}
		if e := validateRequired(r.MaCode, "Mã code", lineNo); e != "" {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return errs
	}

	now := time.Now()
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for i, r := range rows {
			dept := &models.Department{
				Name:      strings.TrimSpace(r.Ten),
				Code:      strings.TrimSpace(r.MaCode),
				Address:   strings.TrimSpace(r.MoTa),
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.deptRepo.CreateInTx(tx, dept); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}
		}
		return nil
	})
	if txErr != nil {
		return []string{txErr.Error()}
	}
	return nil
}

// --- Staff import ---

var validStaffRoles = map[string]models.UserRole{
	"staff":       models.UserRoleStaff,
	"cán bộ":      models.UserRoleStaff,
	"manager":     models.UserRoleManager,
	"quản lý":     models.UserRoleManager,
	"super_admin": models.UserRoleSuperAdmin,
}

func (s *ImportExportService) ImportStaff(rows []StaffImportRow, createdBy string) []string {
	var errs []string
	for i, r := range rows {
		lineNo := i + 2
		if e := validateRequired(r.HoTen, "Họ tên", lineNo); e != "" {
			errs = append(errs, e)
		}
		if e := validateRequired(r.Email, "Email", lineNo); e != "" {
			errs = append(errs, e)
		} else if !emailRegex.MatchString(strings.TrimSpace(r.Email)) {
			errs = append(errs, fmt.Sprintf("Dòng %d: Email '%s' không đúng định dạng", lineNo, r.Email))
		}
		if e := validateRequired(r.VaiTro, "Vai trò", lineNo); e != "" {
			errs = append(errs, e)
		}
		if r.VaiTro != "" {
			if _, ok := validStaffRoles[strings.ToLower(strings.TrimSpace(r.VaiTro))]; !ok {
				errs = append(errs, fmt.Sprintf("Dòng %d: Vai trò '%s' không hợp lệ (staff/manager/super_admin)", lineNo, r.VaiTro))
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}

	now := time.Now()
	defaultHash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword()), bcrypt.DefaultCost)
	if err != nil {
		return []string{"Lỗi hệ thống: không thể tạo mật khẩu mặc định"}
	}

	// Pre-resolve department IDs outside the transaction to avoid nested queries inside TX.
	type staffBatch struct {
		row          StaffImportRow
		role         models.UserRole
		departmentID *string
	}
	batches := make([]staffBatch, len(rows))
	for i, r := range rows {
		b := staffBatch{
			row:  r,
			role: validStaffRoles[strings.ToLower(strings.TrimSpace(r.VaiTro))],
		}
		if code := strings.TrimSpace(r.MaPhongBan); code != "" {
			dept, deptErr := s.deptRepo.FindByCode(context.Background(), code)
			if deptErr != nil {
				return []string{fmt.Sprintf("Dòng %d: lỗi tra cứu phòng ban '%s'", i+2, code)}
			}
			if dept == nil {
				return []string{fmt.Sprintf("Dòng %d: không tìm thấy phòng ban có mã '%s'", i+2, code)}
			}
			b.departmentID = &dept.ID
		}
		batches[i] = b
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for i, b := range batches {
			r := b.row
			user := &models.User{
				Name:         strings.TrimSpace(r.HoTen),
				Email:        strings.TrimSpace(r.Email),
				Phone:        strings.TrimSpace(r.SoDienThoai),
				Role:         b.role,
				Status:       models.UserStatusActive,
				PasswordHash: string(defaultHash),
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.userRepo.CreateInTx(tx, user); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}
			profile := &models.StaffProfile{
				UserID:       user.ID,
				DepartmentID: b.departmentID,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.staffProfileRepo.CreateInTx(tx, profile); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}
		}
		return nil
	})
	if txErr != nil {
		return []string{txErr.Error()}
	}
	return nil
}

// --- ServiceType import ---

func (s *ImportExportService) ImportServiceTypes(rows []ServiceTypeImportRow, createdBy string) []string {
	var errs []string
	for i, r := range rows {
		lineNo := i + 2
		if e := validateRequired(r.Ten, "Tên", lineNo); e != "" {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return errs
	}

	now := time.Now()

	// Pre-resolve department IDs and parse numeric fields outside the transaction.
	type serviceTypeBatch struct {
		row            ServiceTypeImportRow
		processingTime *int
		fee            float64
		departmentID   *string
	}
	batches := make([]serviceTypeBatch, len(rows))
	for i, r := range rows {
		b := serviceTypeBatch{row: r}
		if v := strings.TrimSpace(r.ThoiGianXuLyNgay); v != "" {
			if n, parseErr := strconv.Atoi(v); parseErr == nil {
				b.processingTime = &n
			}
		}
		if v := strings.TrimSpace(r.Phi); v != "" {
			if f, parseErr := strconv.ParseFloat(v, 64); parseErr == nil {
				b.fee = f
			}
		}
		if code := strings.TrimSpace(r.MaPhongBan); code != "" {
			dept, deptErr := s.deptRepo.FindByCode(context.Background(), code)
			if deptErr != nil {
				return []string{fmt.Sprintf("Dòng %d: lỗi tra cứu phòng ban '%s'", i+2, code)}
			}
			if dept == nil {
				return []string{fmt.Sprintf("Dòng %d: không tìm thấy phòng ban có mã '%s'", i+2, code)}
			}
			b.departmentID = &dept.ID
		}
		batches[i] = b
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for i, b := range batches {
			r := b.row
			code := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(r.Ten), " ", "_"))
			if utf8.RuneCountInString(code) > 100 {
				code = string([]rune(code)[:100])
			}
			st := &models.ServiceType{
				Name:                    strings.TrimSpace(r.Ten),
				Code:                    code,
				Description:             strings.TrimSpace(r.MoTa),
				ProcessingTime:          b.processingTime,
				Fee:                     b.fee,
				ResponsibleDepartmentID: b.departmentID,
				FormSchema:              json.RawMessage("{}"),
				IsActive:                true,
				CreatedAt:               now,
				UpdatedAt:               now,
			}
			if err := s.serviceRepo.CreateInTx(tx, st); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}
		}
		return nil
	})
	if txErr != nil {
		return []string{txErr.Error()}
	}
	return nil
}

// --- Citizen import ---

func (s *ImportExportService) ImportCitizens(rows []CitizenImportRow, createdBy string) []string {
	var errs []string
	for i, r := range rows {
		lineNo := i + 2
		if e := validateRequired(r.SoCCCD, "Số CCCD", lineNo); e != "" {
			errs = append(errs, e)
		} else if !cccdRegex.MatchString(strings.TrimSpace(r.SoCCCD)) {
			errs = append(errs, fmt.Sprintf("Dòng %d: Số CCCD phải có đúng 12 chữ số", lineNo))
		}
		if e := validateRequired(r.HoTen, "Họ tên", lineNo); e != "" {
			errs = append(errs, e)
		}
		if e := validateRequired(r.Email, "Email", lineNo); e != "" {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return errs
	}

	now := time.Now()

	// Pre-compute bcrypt hashes outside the transaction to avoid holding DB connections
	// for the ~100ms bcrypt cost per row.
	type citizenBatch struct {
		row          CitizenImportRow
		cccd         string
		passwordHash string
	}
	batches := make([]citizenBatch, len(rows))
	for i, r := range rows {
		cccd := strings.TrimSpace(r.SoCCCD)
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(cccd), bcrypt.DefaultCost)
		if hashErr != nil {
			return []string{fmt.Sprintf("Dòng %d: lỗi tạo mật khẩu", i+2)}
		}
		batches[i] = citizenBatch{row: r, cccd: cccd, passwordHash: string(hash)}
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for i, b := range batches {
			r := b.row
			user := &models.User{
				Name:         strings.TrimSpace(r.HoTen),
				Email:        strings.TrimSpace(r.Email),
				Phone:        strings.TrimSpace(r.SoDienThoai),
				Address:      strings.TrimSpace(r.DiaChi),
				Role:         models.UserRoleCitizen,
				Status:       models.UserStatusActive,
				PasswordHash: b.passwordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.userRepo.CreateInTx(tx, user); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}

			var dob *time.Time
			if trimmed := strings.TrimSpace(r.NgaySinh); trimmed != "" {
				parsed, parseErr := time.Parse("02/01/2006", trimmed)
				if parseErr != nil {
					parsed, parseErr = time.Parse("2006-01-02", trimmed)
				}
				if parseErr == nil {
					dob = &parsed
				}
			}

			profile := &models.CitizenProfile{
				UserID:                   user.ID,
				CitizenIDNumber:          b.cccd,
				DateOfBirth:              dob,
				PermanentAddress:         strings.TrimSpace(r.DiaChi),
				EmailNotificationEnabled: true,
				CreatedAt:                now,
				UpdatedAt:                now,
			}
			if err := s.profileRepo.CreateInTx(tx, profile); err != nil {
				return fmt.Errorf("Dòng %d: %w", i+2, err)
			}
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrDuplicatedKey) {
			return []string{"Dữ liệu bị trùng (email hoặc số CCCD đã tồn tại): " + txErr.Error()}
		}
		return []string{txErr.Error()}
	}
	return nil
}
