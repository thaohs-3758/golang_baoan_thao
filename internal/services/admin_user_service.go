package services

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

func defaultAdminPassword() string {
	if p := os.Getenv("DEFAULT_ADMIN_PASSWORD"); p != "" {
		return p
	}
	return "Aa@123456"
}

var ErrUserNotFoundAdmin = errors.New("admin.user_not_found")

type AdminUserService struct {
	userRepo       repositories.UserRepository
	activityLogger activityLogger
}

func NewAdminUserService(userRepo repositories.UserRepository, loggers ...activityLogger) *AdminUserService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &AdminUserService{userRepo: userRepo, activityLogger: logger}
}

func (s *AdminUserService) ListUsers(filter repositories.UserFilter, page, limit int) ([]models.User, int64, error) {
	offset := (page - 1) * limit
	return s.userRepo.List(filter, offset, limit)
}

func (s *AdminUserService) GetUser(id string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFoundAdmin
	}
	return user, nil
}

func (s *AdminUserService) CreateUser(req *dtos.AdminCreateUserRequest, createdBy string) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	password := req.Password
	if password == "" {
		password = defaultAdminPassword()
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.UserRole(req.Role),
		Phone:        req.Phone,
		Address:      req.Address,
		Status:       models.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &createdBy,
	}
	created, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &createdBy,
		Action:      "user.create",
		EntityType:  "user",
		EntityID:    &created.ID,
		Description: "Tạo người dùng: " + created.Name + " (" + created.Email + ")",
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"role":   created.Role,
			"status": created.Status,
		}),
		CreatedAt: now,
	})
	return created, nil
}

func (s *AdminUserService) UpdateUser(id string, req *dtos.AdminUpdateUserRequest, updatedBy string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFoundAdmin
	}
	prevName := user.Name
	prevRole := user.Role
	prevPhone := user.Phone
	prevAddress := user.Address

	user.Name = req.Name
	user.Role = models.UserRole(req.Role)
	user.Phone = req.Phone
	user.Address = req.Address
	user.UpdatedAt = time.Now()
	user.UpdatedBy = &updatedBy

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &updatedBy,
		Action:      "user.update",
		EntityType:  "user",
		EntityID:    &user.ID,
		Description: "Cập nhật người dùng: " + user.Name,
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"changes": map[string]any{
				"name":    map[string]any{"before": prevName, "after": req.Name},
				"role":    map[string]any{"before": prevRole, "after": req.Role},
				"phone":   map[string]any{"before": prevPhone, "after": req.Phone},
				"address": map[string]any{"before": prevAddress, "after": req.Address},
			},
		}),
		CreatedAt: user.UpdatedAt,
	})
	return user, nil
}

func (s *AdminUserService) BlockUser(id string, updatedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	if err := s.userRepo.UpdateStatus(id, models.UserStatusBlocked, updatedBy); err != nil {
		return err
	}
	now := time.Now()
	s.logActivity(&models.ActivityLog{
		ActorUserID: &updatedBy,
		Action:      "user.block",
		EntityType:  "user",
		EntityID:    &id,
		Description: "Khóa tài khoản: " + user.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return nil
}

func (s *AdminUserService) UnblockUser(id string, updatedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	if err := s.userRepo.UpdateStatus(id, models.UserStatusActive, updatedBy); err != nil {
		return err
	}
	now := time.Now()
	s.logActivity(&models.ActivityLog{
		ActorUserID: &updatedBy,
		Action:      "user.unblock",
		EntityType:  "user",
		EntityID:    &id,
		Description: "Mở khóa tài khoản: " + user.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return nil
}

func (s *AdminUserService) DeleteUser(id string, deletedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	if err := s.userRepo.SoftDelete(id, deletedBy); err != nil {
		return err
	}
	now := time.Now()
	s.logActivity(&models.ActivityLog{
		ActorUserID: &deletedBy,
		Action:      "user.delete",
		EntityType:  "user",
		EntityID:    &id,
		Description: "Xóa người dùng: " + user.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return nil
}

func (s *AdminUserService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}
