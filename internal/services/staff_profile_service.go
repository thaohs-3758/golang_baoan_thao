package services

import (
	"context"
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrLeaderTransferForbiddenForManager = errors.New("department.leader_transfer_forbidden_for_manager")

type StaffProfileService struct {
	repo     repositories.StaffProfileRepository
	userRepo repositories.UserRepository
	deptRepo repositories.DepartmentRepository
}

func NewStaffProfileService(repo repositories.StaffProfileRepository, userRepo repositories.UserRepository, deptRepo ...repositories.DepartmentRepository) *StaffProfileService {
	svc := &StaffProfileService{repo: repo, userRepo: userRepo}
	if len(deptRepo) > 0 {
		svc.deptRepo = deptRepo[0]
	}
	return svc
}

func (s *StaffProfileService) ListStaffByDepartment(deptID string, page, limit int) ([]models.StaffProfile, int64, error) {
	offset := (page - 1) * limit
	return s.repo.ListByDepartment(deptID, offset, limit)
}

func (s *StaffProfileService) FindStaffProfileByUserID(userID string) (*models.StaffProfile, error) {
	return s.repo.FindByUserID(userID)
}

func (s *StaffProfileService) AssignStaffToDepartment(userID string, deptID string, updatedBy string) error {
	// validate user exists
	u, err := s.userRepo.FindByID(userID)
	if err != nil || u == nil {
		return err
	}
	actor, err := s.userRepo.FindByID(updatedBy)
	if err != nil || actor == nil {
		return err
	}
	if actor.Role == models.UserRoleManager && s.deptRepo != nil {
		leaderDept, err := s.deptRepo.FindByLeaderUserID(context.Background(), userID)
		if err != nil {
			return err
		}
		if leaderDept != nil && leaderDept.ID != deptID {
			return ErrLeaderTransferForbiddenForManager
		}
	}
	return s.repo.UpdateDepartment(userID, &deptID, updatedBy)
}

func (s *StaffProfileService) RemoveStaffFromDepartment(userID string, updatedBy string) error {
	actor, err := s.userRepo.FindByID(updatedBy)
	if err != nil || actor == nil {
		return err
	}
	if actor.Role == models.UserRoleManager && s.deptRepo != nil {
		leaderDept, err := s.deptRepo.FindByLeaderUserID(context.Background(), userID)
		if err != nil {
			return err
		}
		if leaderDept != nil {
			return ErrLeaderTransferForbiddenForManager
		}
	}
	return s.repo.UpdateDepartment(userID, nil, updatedBy)
}
