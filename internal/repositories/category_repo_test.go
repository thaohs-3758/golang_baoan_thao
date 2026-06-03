package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockCategoryRepo(t *testing.T) (CategoryRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}
	return NewCategoryRepo(db), mock, func() { _ = sqlDB.Close() }
}

func TestCategoryRepo_FindByID_Found(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow("cat-1", "Test Category", "TEST")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE id = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("cat-1", 1).
		WillReturnRows(rows)

	cat, err := repo.FindByID(context.Background(), "cat-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat == nil || cat.ID != "cat-1" {
		t.Fatalf("expected category cat-1, got %v", cat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_FindByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE id = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	cat, err := repo.FindByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != nil {
		t.Fatal("expected nil cat for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_FindByID_Error(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE id = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("cat-err", 1).
		WillReturnError(dbErr)

	cat, err := repo.FindByID(context.Background(), "cat-err")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if cat != nil {
		t.Fatal("expected nil cat on error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_FindByCode_Found(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow("cat-1", "Test Category", "TEST")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE code = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("TEST", 1).
		WillReturnRows(rows)

	cat, err := repo.FindByCode(context.Background(), "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat == nil || cat.Code != "TEST" {
		t.Fatalf("expected code TEST, got %v", cat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_FindByCode_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE code = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("MISSING", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	cat, err := repo.FindByCode(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != nil {
		t.Fatal("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_FindByCode_Error(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	dbErr := errors.New("db connection error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE code = $1 AND deleted_at IS NULL ORDER BY "categories"."id" LIMIT $2`)).
		WithArgs("CODE", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByCode(context.Background(), "CODE")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_Create(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	cat := &models.Category{ID: "cat-new", Name: "New", Code: "NEW", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categories"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("cat-new"))
	mock.ExpectCommit()

	result, err := repo.Create(context.Background(), cat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "cat-new" {
		t.Fatalf("expected created category, got %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	cat := &models.Category{ID: "cat-fail", Name: "Fail", Code: "FAIL", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categories"`)).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	result, err := repo.Create(context.Background(), cat)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatal("expected nil result on error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_Update(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	cat := &models.Category{ID: "cat-1", Name: "Updated", Code: "UPD", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "categories"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), cat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_List_NoFilter(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categories" WHERE deleted_at IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	rows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow("cat-1", "Cat 1", "CAT1").
		AddRow("cat-2", "Cat 2", "CAT2")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1`)).
		WithArgs(10).
		WillReturnRows(rows)

	cats, total, err := repo.List(context.Background(), CategoryFilter{}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_List_WithSearch(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categories" WHERE deleted_at IS NULL AND (LOWER(name) LIKE $1 OR LOWER(code) LIKE $2)`)).
		WithArgs("%test%", "%test%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow("cat-1", "Test Cat", "TEST")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE deleted_at IS NULL AND (LOWER(name) LIKE $1 OR LOWER(code) LIKE $2) ORDER BY created_at DESC LIMIT $3`)).
		WithArgs("%test%", "%test%", 10).
		WillReturnRows(rows)

	cats, total, err := repo.List(context.Background(), CategoryFilter{Search: "test"}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_List_CountError(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "categories"`).WillReturnError(dbErr)

	_, _, err := repo.List(context.Background(), CategoryFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_List_FindError(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "categories"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT \* FROM "categories"`).WillReturnError(dbErr)

	_, _, err := repo.List(context.Background(), CategoryFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoryRepo_SoftDelete(t *testing.T) {
	repo, mock, cleanup := newMockCategoryRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "categories" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "cat-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SoftDelete(context.Background(), "cat-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
