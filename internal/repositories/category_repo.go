package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type CategoryFilter struct {
	Search string
}

type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (*models.Category, error)
	FindByCode(ctx context.Context, code string) (*models.Category, error)
	Create(ctx context.Context, cat *models.Category) (*models.Category, error)
	Update(ctx context.Context, cat *models.Category) error
	List(ctx context.Context, filter CategoryFilter, offset, limit int) ([]models.Category, int64, error)
	SoftDelete(ctx context.Context, id string, deletedBy string) error
}

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) CategoryRepository {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) FindByID(ctx context.Context, id string) (*models.Category, error) {
	var cat models.Category
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) FindByCode(ctx context.Context, code string) (*models.Category, error) {
	var cat models.Category
	err := r.db.WithContext(ctx).Where("code = ? AND deleted_at IS NULL", code).First(&cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) Create(ctx context.Context, cat *models.Category) (*models.Category, error) {
	if err := r.db.WithContext(ctx).Create(cat).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

func (r *CategoryRepo) Update(ctx context.Context, cat *models.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *CategoryRepo) List(ctx context.Context, filter CategoryFilter, offset, limit int) ([]models.Category, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Category{}).Where("deleted_at IS NULL")
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var cats []models.Category
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&cats).Error; err != nil {
		return nil, 0, err
	}
	return cats, total, nil
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
			"deleted_by": deletedBy,
		}).Error
}
