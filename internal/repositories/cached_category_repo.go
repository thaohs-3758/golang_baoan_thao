package repositories

import (
	"context"

	appcache "github.com/awesome-academy/golang_baoan_thao/internal/cache"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

type cachedCategoryRepo struct {
	base  CategoryRepository
	cache *appcache.JSONCache
}

func NewCachedCategoryRepository(base CategoryRepository, cache *appcache.JSONCache) CategoryRepository {
	return &cachedCategoryRepo{base: base, cache: cache}
}

func (r *cachedCategoryRepo) FindByID(ctx context.Context, id string) (*models.Category, error) {
	key := "cache:categories:detail:" + id
	var cached models.Category
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

func (r *cachedCategoryRepo) FindByCode(ctx context.Context, code string) (*models.Category, error) {
	key := "cache:categories:detail:code:" + code
	var cached models.Category
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

func (r *cachedCategoryRepo) Create(ctx context.Context, cat *models.Category) (*models.Category, error) {
	result, err := r.base.Create(ctx, cat)
	if err != nil {
		return nil, err
	}
	r.cache.DeletePattern(ctx, "cache:categories:*")
	return result, nil
}

func (r *cachedCategoryRepo) Update(ctx context.Context, cat *models.Category) error {
	if err := r.base.Update(ctx, cat); err != nil {
		return err
	}
	r.cache.DeletePattern(ctx, "cache:categories:*")
	return nil
}

func (r *cachedCategoryRepo) List(ctx context.Context, filter CategoryFilter, offset, limit int) ([]models.Category, int64, error) {
	key := "cache:categories:list:search=" + filter.Search
	var cached []models.Category
	if r.cache.Get(ctx, key, &cached) {
		return cached, int64(len(cached)), nil
	}
	result, _, err := r.base.List(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	r.cache.Set(ctx, key, result)
	return result, int64(len(result)), nil
}

func (r *cachedCategoryRepo) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	if err := r.base.SoftDelete(ctx, id, deletedBy); err != nil {
		return err
	}
	r.cache.DeletePattern(ctx, "cache:categories:*")
	return nil
}
