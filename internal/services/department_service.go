package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrDepartmentNotFound = errors.New("department.not_found")
var ErrDepartmentCodeExists = errors.New("department.code_exists")
var ErrDepartmentLeaderAlreadyAssigned = errors.New("department.leader_already_assigned")

type DepartmentService struct {
	repo           repositories.DepartmentRepository
	staffRepo      repositories.StaffProfileRepository
	activityLogger activityLogger
}

func NewDepartmentService(repo repositories.DepartmentRepository, staffRepo repositories.StaffProfileRepository, loggers ...activityLogger) *DepartmentService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &DepartmentService{repo: repo, staffRepo: staffRepo, activityLogger: logger}
}

func (s *DepartmentService) ListDepartments(filter repositories.DepartmentFilter, page, limit int) ([]models.Department, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(context.Background(), filter, offset, limit)
}

func (s *DepartmentService) GetDepartment(id string) (*models.Department, error) {
	dept, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}
	return dept, nil
}

func (s *DepartmentService) CreateDepartment(req *dtos.DepartmentCreateRequest, createdBy string) (*models.Department, error) {
	existing, err := s.repo.FindByCode(context.Background(), req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDepartmentCodeExists
	}

	now := time.Now()
	dept := &models.Department{
		Name:      req.Name,
		Code:      req.Code,
		Address:   req.Address,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.LeaderUserID != "" {
		leaderDept, err := s.repo.FindByLeaderUserID(context.Background(), req.LeaderUserID)
		if err != nil {
			return nil, err
		}
		if leaderDept != nil {
			return nil, ErrDepartmentLeaderAlreadyAssigned
		}
		dept.LeaderUserID = &req.LeaderUserID
	}

	created, err := s.repo.Create(context.Background(), dept)
	if err != nil {
		return nil, err
	}
	if created.LeaderUserID != nil && *created.LeaderUserID != "" && s.staffRepo != nil {
		// sync leader profile after department has a real ID
		if err := s.staffRepo.UpdateDepartment(*created.LeaderUserID, &created.ID, createdBy); err != nil {
			return nil, err
		}
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &createdBy,
		Action:      "department.create",
		EntityType:  "department",
		EntityID:    &created.ID,
		Description: "Tạo phòng ban: " + created.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return created, nil
}

func (s *DepartmentService) UpdateDepartment(id string, req *dtos.DepartmentUpdateRequest, updatedBy string) (*models.Department, error) {
	dept, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}
	prevName := dept.Name
	prevCode := dept.Code
	prevAddress := dept.Address

	if req.Code != dept.Code {
		existing, err := s.repo.FindByCode(context.Background(), req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrDepartmentCodeExists
		}
	}

	dept.Name = req.Name
	dept.Code = req.Code
	dept.Address = req.Address
	dept.UpdatedAt = time.Now()

	prevLeaderID := dept.LeaderUserID
	if req.LeaderUserID != "" {
		leaderDept, err := s.repo.FindByLeaderUserID(context.Background(), req.LeaderUserID)
		if err != nil {
			return nil, err
		}
		if leaderDept != nil && leaderDept.ID != dept.ID {
			return nil, ErrDepartmentLeaderAlreadyAssigned
		}
		dept.LeaderUserID = &req.LeaderUserID
	} else {
		dept.LeaderUserID = nil
	}

	if err := s.repo.Update(context.Background(), dept); err != nil {
		return nil, err
	}
	if s.staffRepo != nil {
		newLeaderID := ""
		if dept.LeaderUserID != nil {
			newLeaderID = *dept.LeaderUserID
		}
		if prevLeaderID != nil && *prevLeaderID != "" && *prevLeaderID != newLeaderID {
			if err := s.staffRepo.UpdateDepartment(*prevLeaderID, nil, updatedBy); err != nil {
				return nil, err
			}
		}
		if newLeaderID != "" {
			if err := s.staffRepo.UpdateDepartment(newLeaderID, &dept.ID, updatedBy); err != nil {
				return nil, err
			}
		}
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &updatedBy,
		Action:      "department.update",
		EntityType:  "department",
		EntityID:    &dept.ID,
		Description: "Cập nhật phòng ban: " + dept.Name,
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"changes": map[string]any{
				"name":    map[string]any{"before": prevName, "after": req.Name},
				"code":    map[string]any{"before": prevCode, "after": req.Code},
				"address": map[string]any{"before": prevAddress, "after": req.Address},
			},
		}),
		CreatedAt: dept.UpdatedAt,
	})
	return dept, nil
}

func (s *DepartmentService) DeleteDepartment(id string, deletedBy string) error {
	dept, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	if dept == nil {
		return ErrDepartmentNotFound
	}
	if err := s.repo.SoftDelete(context.Background(), id, deletedBy); err != nil {
		return err
	}
	now := time.Now()
	s.logActivity(&models.ActivityLog{
		ActorUserID: &deletedBy,
		Action:      "department.delete",
		EntityType:  "department",
		EntityID:    &id,
		Description: "Xóa phòng ban: " + dept.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return nil
}

func (s *DepartmentService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}
