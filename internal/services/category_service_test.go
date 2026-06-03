package services

import (
	"context"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// --- fakeCategoryRepo ---

type fakeCategoryRepo struct {
	cat       *models.Category
	cats      []models.Category
	total     int64
	findErr   error
	codeErr   error
	codeCat   *models.Category
	createErr error
	updateErr error
	deleteErr error
}

func (r *fakeCategoryRepo) FindByID(_ context.Context, _ string) (*models.Category, error) {
	return r.cat, r.findErr
}
func (r *fakeCategoryRepo) FindByCode(_ context.Context, _ string) (*models.Category, error) {
	return r.codeCat, r.codeErr
}
func (r *fakeCategoryRepo) Create(_ context.Context, c *models.Category) (*models.Category, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return c, nil
}
func (r *fakeCategoryRepo) Update(_ context.Context, _ *models.Category) error { return r.updateErr }
func (r *fakeCategoryRepo) List(_ context.Context, _ repositories.CategoryFilter, _, _ int) ([]models.Category, int64, error) {
	if r.findErr != nil {
		return nil, 0, r.findErr
	}
	return r.cats, r.total, nil
}
func (r *fakeCategoryRepo) SoftDelete(_ context.Context, _ string, _ string) error {
	return r.deleteErr
}

var _ repositories.CategoryRepository = (*fakeCategoryRepo)(nil)

func newCatSvc(repo *fakeCategoryRepo) *CategoryService {
	return NewCategoryService(repo)
}

// --- ListCategories ---

func TestCategoryService_ListCategories_OK(t *testing.T) {
	cats := []models.Category{{ID: "1", Name: "Hành chính"}}
	svc := newCatSvc(&fakeCategoryRepo{cats: cats, total: 1})
	result, total, err := svc.ListCategories(repositories.CategoryFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestCategoryService_ListCategories_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newCatSvc(&fakeCategoryRepo{findErr: repoErr})
	_, _, err := svc.ListCategories(repositories.CategoryFilter{}, 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetCategory ---

func TestCategoryService_GetCategory_Found(t *testing.T) {
	c := &models.Category{ID: "c1", Name: "Y tế"}
	svc := newCatSvc(&fakeCategoryRepo{cat: c})
	result, err := svc.GetCategory("c1")
	assert.NoError(t, err)
	assert.Equal(t, "c1", result.ID)
}

func TestCategoryService_GetCategory_NotFound(t *testing.T) {
	svc := newCatSvc(&fakeCategoryRepo{cat: nil})
	_, err := svc.GetCategory("missing")
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryService_GetCategory_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newCatSvc(&fakeCategoryRepo{findErr: repoErr})
	_, err := svc.GetCategory("x")
	assert.ErrorIs(t, err, repoErr)
}

// --- CreateCategory ---

func TestCategoryService_CreateCategory_OK(t *testing.T) {
	svc := newCatSvc(&fakeCategoryRepo{codeCat: nil})
	req := &dtos.CategoryCreateRequest{Name: "Y tế", Code: "y_te", Description: "Lĩnh vực y tế"}
	cat, err := svc.CreateCategory(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "Y tế", cat.Name)
	assert.Equal(t, "y_te", cat.Code)
	assert.True(t, cat.IsActive)
}

func TestCategoryService_CreateCategory_CodeExists(t *testing.T) {
	existing := &models.Category{Code: "y_te"}
	svc := newCatSvc(&fakeCategoryRepo{codeCat: existing})
	req := &dtos.CategoryCreateRequest{Name: "Y tế", Code: "y_te"}
	_, err := svc.CreateCategory(req, "actor")
	assert.ErrorIs(t, err, ErrCategoryCodeExists)
}

func TestCategoryService_CreateCategory_FindCodeError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newCatSvc(&fakeCategoryRepo{codeErr: repoErr})
	req := &dtos.CategoryCreateRequest{Name: "Y tế", Code: "y_te"}
	_, err := svc.CreateCategory(req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestCategoryService_CreateCategory_CreateError(t *testing.T) {
	createErr := errors.New("create failed")
	svc := newCatSvc(&fakeCategoryRepo{createErr: createErr})
	req := &dtos.CategoryCreateRequest{Name: "Y tế", Code: "y_te"}
	_, err := svc.CreateCategory(req, "actor")
	assert.ErrorIs(t, err, createErr)
}

func TestCategoryService_CreateCategory_DBUniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	svc := newCatSvc(&fakeCategoryRepo{createErr: pgErr})
	req := &dtos.CategoryCreateRequest{Name: "Y tế", Code: "y_te"}
	_, err := svc.CreateCategory(req, "actor")
	assert.ErrorIs(t, err, ErrCategoryCodeExists)
}

// --- UpdateCategory ---

func TestCategoryService_UpdateCategory_OK(t *testing.T) {
	c := &models.Category{ID: "c1", Name: "Old", Code: "old_code", IsActive: true}
	svc := newCatSvc(&fakeCategoryRepo{cat: c})
	req := &dtos.CategoryUpdateRequest{Name: "New", Code: "old_code", IsActive: false}
	result, err := svc.UpdateCategory("c1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.False(t, result.IsActive)
}

func TestCategoryService_UpdateCategory_CodeChange_Unique(t *testing.T) {
	c := &models.Category{ID: "c1", Code: "old"}
	svc := newCatSvc(&fakeCategoryRepo{cat: c, codeCat: nil})
	req := &dtos.CategoryUpdateRequest{Name: "Cat", Code: "new"}
	result, err := svc.UpdateCategory("c1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "new", result.Code)
}

func TestCategoryService_UpdateCategory_CodeChange_Conflict(t *testing.T) {
	c := &models.Category{ID: "c1", Code: "old"}
	existing := &models.Category{ID: "c2", Code: "new"}
	svc := newCatSvc(&fakeCategoryRepo{cat: c, codeCat: existing})
	req := &dtos.CategoryUpdateRequest{Name: "Cat", Code: "new"}
	_, err := svc.UpdateCategory("c1", req, "actor")
	assert.ErrorIs(t, err, ErrCategoryCodeExists)
}

func TestCategoryService_UpdateCategory_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newCatSvc(&fakeCategoryRepo{findErr: repoErr})
	req := &dtos.CategoryUpdateRequest{Name: "X", Code: "x"}
	_, err := svc.UpdateCategory("c1", req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestCategoryService_UpdateCategory_CodeChange_FindCodeError(t *testing.T) {
	c := &models.Category{ID: "c1", Code: "old"}
	codeErr := errors.New("code lookup failed")
	svc := newCatSvc(&fakeCategoryRepo{cat: c, codeErr: codeErr})
	req := &dtos.CategoryUpdateRequest{Name: "Cat", Code: "new"}
	_, err := svc.UpdateCategory("c1", req, "actor")
	assert.ErrorIs(t, err, codeErr)
}

func TestCategoryService_UpdateCategory_NotFound(t *testing.T) {
	svc := newCatSvc(&fakeCategoryRepo{cat: nil})
	req := &dtos.CategoryUpdateRequest{Name: "X", Code: "x"}
	_, err := svc.UpdateCategory("missing", req, "actor")
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryService_UpdateCategory_SaveError(t *testing.T) {
	c := &models.Category{ID: "c1", Code: "old"}
	saveErr := errors.New("save failed")
	svc := newCatSvc(&fakeCategoryRepo{cat: c, updateErr: saveErr})
	req := &dtos.CategoryUpdateRequest{Name: "X", Code: "old"}
	_, err := svc.UpdateCategory("c1", req, "actor")
	assert.ErrorIs(t, err, saveErr)
}

func TestCategoryService_UpdateCategory_DBUniqueViolation(t *testing.T) {
	c := &models.Category{ID: "c1", Code: "old"}
	pgErr := &pgconn.PgError{Code: "23505"}
	svc := newCatSvc(&fakeCategoryRepo{cat: c, updateErr: pgErr})
	req := &dtos.CategoryUpdateRequest{Name: "X", Code: "old"}
	_, err := svc.UpdateCategory("c1", req, "actor")
	assert.ErrorIs(t, err, ErrCategoryCodeExists)
}

// --- DeleteCategory ---

func TestCategoryService_DeleteCategory_OK(t *testing.T) {
	c := &models.Category{ID: "c1"}
	svc := newCatSvc(&fakeCategoryRepo{cat: c})
	err := svc.DeleteCategory("c1", "actor")
	assert.NoError(t, err)
}

func TestCategoryService_DeleteCategory_NotFound(t *testing.T) {
	svc := newCatSvc(&fakeCategoryRepo{cat: nil})
	err := svc.DeleteCategory("missing", "actor")
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryService_DeleteCategory_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newCatSvc(&fakeCategoryRepo{findErr: repoErr})
	err := svc.DeleteCategory("c1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestCategoryService_DeleteCategory_DeleteError(t *testing.T) {
	c := &models.Category{ID: "c1"}
	deleteErr := errors.New("delete error")
	svc := newCatSvc(&fakeCategoryRepo{cat: c, deleteErr: deleteErr})
	err := svc.DeleteCategory("c1", "actor")
	assert.ErrorIs(t, err, deleteErr)
}

func TestCategoryService_LogActivity_WithLogger(t *testing.T) {
	cat := &models.Category{ID: "c1", Name: "Hành chính"}
	repo := &fakeCategoryRepo{codeCat: nil, cat: cat}
	logger := &fakeActivityLogger{}
	svc := NewCategoryService(repo, logger)

	req := &dtos.CategoryCreateRequest{Name: "Test", Code: "TEST"}
	_, err := svc.CreateCategory(req, "actor1")
	assert.NoError(t, err)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "category.create", logger.entries[0].Action)
	}
}

func TestCategoryService_LogActivity_LoggerError(t *testing.T) {
	cat := &models.Category{ID: "c1", Name: "Hành chính"}
	repo := &fakeCategoryRepo{codeCat: nil, cat: cat}
	logger := &fakeActivityLogger{err: assert.AnError}
	svc := NewCategoryService(repo, logger)

	req := &dtos.CategoryCreateRequest{Name: "Test", Code: "TEST"}
	// Should succeed even if logger fails
	_, err := svc.CreateCategory(req, "actor1")
	assert.NoError(t, err)
}
