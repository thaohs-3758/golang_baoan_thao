package repositories

import (
	"context"

	appcache "github.com/awesome-academy/golang_baoan_thao/internal/cache"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type cachedDepartmentRepo struct {
	base  DepartmentRepository
	cache *appcache.JSONCache
}

func NewCachedDepartmentRepository(base DepartmentRepository, cache *appcache.JSONCache) DepartmentRepository {
	return &cachedDepartmentRepo{base: base, cache: cache}
}

func (r *cachedDepartmentRepo) FindByID(ctx context.Context, id string) (*models.Department, error) {
	key := "cache:departments:detail:" + id
	var cached models.Department
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}
	result, err := r.base.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedDepartmentRepo) FindByCode(ctx context.Context, code string) (*models.Department, error) {
	key := "cache:departments:detail:code:" + code
	var cached models.Department
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}
	result, err := r.base.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedDepartmentRepo) FindByLeaderUserID(ctx context.Context, userID string) (*models.Department, error) {
	key := "cache:departments:detail:leader_user_id:" + userID
	var cached models.Department
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}
	result, err := r.base.FindByLeaderUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedDepartmentRepo) Create(ctx context.Context, dept *models.Department) (*models.Department, error) {
	result, err := r.base.Create(ctx, dept)
	if err != nil {
		return nil, err
	}
	r.cache.DeletePattern(ctx, "cache:departments:*")
	return result, nil
}

func (r *cachedDepartmentRepo) Update(ctx context.Context, dept *models.Department) error {
	if err := r.base.Update(ctx, dept); err != nil {
		return err
	}
	r.cache.DeletePattern(ctx, "cache:departments:*")
	return nil
}

func (r *cachedDepartmentRepo) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	if err := r.base.SoftDelete(ctx, id, deletedBy); err != nil {
		return err
	}
	r.cache.DeletePattern(ctx, "cache:departments:*")
	return nil
}

func (r *cachedDepartmentRepo) List(ctx context.Context, filter DepartmentFilter, offset, limit int) ([]models.Department, int64, error) {
	key := "cache:departments:list:search=" + filter.Search
	var cached []models.Department
	if r.cache.Get(ctx, key, &cached) {
		return cached, int64(len(cached)), nil
	}
	result, count, err := r.base.List(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	r.cache.Set(ctx, key, result)
	return result, count, nil
}

func (r *cachedDepartmentRepo) CreateInTx(tx *gorm.DB, dept *models.Department) error {
	return r.base.CreateInTx(tx, dept)
}
