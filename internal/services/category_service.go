package services

import (
	"errors"
	"log"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/jackc/pgx/v5/pgconn"
)

func isCategoryUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var ErrCategoryNotFound = errors.New("category.not_found")
var ErrCategoryCodeExists = errors.New("category.code_exists")

type CategoryService struct {
	repo           repositories.CategoryRepository
	activityLogger activityLogger
}

func NewCategoryService(repo repositories.CategoryRepository, loggers ...activityLogger) *CategoryService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &CategoryService{repo: repo, activityLogger: logger}
}

func (s *CategoryService) ListCategories(filter repositories.CategoryFilter, page, limit int) ([]models.Category, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(filter, offset, limit)
}

func (s *CategoryService) GetCategory(id string) (*models.Category, error) {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}
	return cat, nil
}

func (s *CategoryService) CreateCategory(req *dtos.CategoryCreateRequest, createdBy string) (*models.Category, error) {
	existing, err := s.repo.FindByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCategoryCodeExists
	}

	now := time.Now()
	cat := &models.Category{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.repo.Create(cat)
	if err != nil {
		if isCategoryUniqueViolation(err) {
			return nil, ErrCategoryCodeExists
		}
		return nil, err
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &createdBy,
		Action:      "category.create",
		EntityType:  "category",
		EntityID:    &created.ID,
		Description: "Tạo danh mục: " + created.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return created, nil
}

func (s *CategoryService) UpdateCategory(id string, req *dtos.CategoryUpdateRequest, updatedBy string) (*models.Category, error) {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}
	prevName := cat.Name
	prevCode := cat.Code
	prevDescription := cat.Description
	prevActive := cat.IsActive

	if req.Code != cat.Code {
		existing, err := s.repo.FindByCode(req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrCategoryCodeExists
		}
	}

	cat.Name = req.Name
	cat.Code = req.Code
	cat.Description = req.Description
	cat.IsActive = req.IsActive
	cat.UpdatedAt = time.Now()

	if err := s.repo.Update(cat); err != nil {
		if isCategoryUniqueViolation(err) {
			return nil, ErrCategoryCodeExists
		}
		return nil, err
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID: &updatedBy,
		Action:      "category.update",
		EntityType:  "category",
		EntityID:    &cat.ID,
		Description: "Cập nhật danh mục: " + cat.Name,
		Result:      "success",
		MetadataJSON: mustJSON(map[string]any{
			"changes": map[string]any{
				"name":        map[string]any{"before": prevName, "after": req.Name},
				"code":        map[string]any{"before": prevCode, "after": req.Code},
				"description": map[string]any{"before": prevDescription, "after": req.Description},
				"is_active":   map[string]any{"before": prevActive, "after": req.IsActive},
			},
		}),
		CreatedAt: cat.UpdatedAt,
	})
	return cat, nil
}

func (s *CategoryService) DeleteCategory(id string, deletedBy string) error {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if cat == nil {
		return ErrCategoryNotFound
	}
	if err := s.repo.SoftDelete(id, deletedBy); err != nil {
		return err
	}
	now := time.Now()
	s.logActivity(&models.ActivityLog{
		ActorUserID: &deletedBy,
		Action:      "category.delete",
		EntityType:  "category",
		EntityID:    &id,
		Description: "Xóa danh mục: " + cat.Name,
		Result:      "success",
		CreatedAt:   now,
	})
	return nil
}

func (s *CategoryService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}
