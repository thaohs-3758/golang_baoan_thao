package repositories

import (
	"context"
	"fmt"

	appcache "github.com/awesome-academy/golang_baoan_thao/internal/cache"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type cachedServiceTypeRepo struct {
	base  ServiceTypeRepository
	cache *appcache.JSONCache
}

func NewCachedServiceTypeRepository(base ServiceTypeRepository, cache *appcache.JSONCache) ServiceTypeRepository {
	return &cachedServiceTypeRepo{base: base, cache: cache}
}

func (r *cachedServiceTypeRepo) List(ctx context.Context, filter ListFilter) (*ListResult, error) {
	key := fmt.Sprintf(
		"cache:service_types:list:category=%s:search=%s:page=%d:limit=%d:inactive=%t",
		filter.Category,
		filter.Search,
		filter.Page,
		filter.Limit,
		filter.IncludeInactive,
	)

	var cached ListResult
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}

	result, err := r.base.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedServiceTypeRepo) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	key := "cache:service_types:detail:" + id + ":active=true"

	var cached models.ServiceType
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}

	result, err := r.base.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedServiceTypeRepo) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	key := "cache:service_types:detail:" + id + ":active=false"

	var cached models.ServiceType
	if r.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}

	result, err := r.base.GetByIDForAdmin(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedServiceTypeRepo) ListDepartments(ctx context.Context) ([]models.Department, error) {
	key := "cache:departments:list"

	var cached []models.Department
	if r.cache.Get(ctx, key, &cached) {
		return cached, nil
	}

	result, err := r.base.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedServiceTypeRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	key := "cache:categories:list:active=true"

	var cached []models.Category
	if r.cache.Get(ctx, key, &cached) {
		return cached, nil
	}

	result, err := r.base.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.Set(ctx, key, result)
	return result, nil
}

func (r *cachedServiceTypeRepo) Create(ctx context.Context, st *models.ServiceType) error {
	if err := r.base.Create(ctx, st); err != nil {
		return err
	}

	r.invalidate(ctx)
	return nil
}

func (r *cachedServiceTypeRepo) Update(ctx context.Context, st *models.ServiceType) error {
	if err := r.base.Update(ctx, st); err != nil {
		return err
	}

	r.invalidate(ctx)
	return nil
}

func (r *cachedServiceTypeRepo) Delete(ctx context.Context, id string) error {
	if err := r.base.Delete(ctx, id); err != nil {
		return err
	}

	r.invalidate(ctx)
	return nil
}

func (r *cachedServiceTypeRepo) CreateInTx(tx *gorm.DB, st *models.ServiceType) error {
	return r.base.CreateInTx(tx, st)
}

func (r *cachedServiceTypeRepo) CountApplications(ctx context.Context, serviceTypeID string) (int64, error) {
	return r.base.CountApplications(ctx, serviceTypeID)
}

func (r *cachedServiceTypeRepo) invalidate(ctx context.Context) {
	r.cache.DeletePattern(ctx, "cache:service_types:*")
	r.cache.DeletePattern(ctx, "cache:categories:*")
	r.cache.DeletePattern(ctx, "cache:departments:*")
}
