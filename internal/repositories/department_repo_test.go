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

func newMockDeptRepo(t *testing.T) (DepartmentRepository, sqlmock.Sqlmock, func()) {
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
	return NewDepartmentRepo(db), mock, func() { _ = sqlDB.Close() }
}

func TestDeptRepo_FindByID_Found(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "code", "address", "leader_user_id", "created_at", "updated_at", "deleted_at"}).
		AddRow("d1", "IT Dept", "IT", "123 St", nil, time.Now(), time.Now(), nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("d1", 1).
		WillReturnRows(rows)

	// Preload LeaderUser (no leader_user_id so no preload query needed, but GORM still issues it)
	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	dept, err := repo.FindByID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept == nil || dept.ID != "d1" {
		t.Fatalf("expected dept d1, got %#v", dept)
	}
}

func TestDeptRepo_FindByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	dept, err := repo.FindByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept != nil {
		t.Fatalf("expected nil dept, got %#v", dept)
	}
}

func TestDeptRepo_FindByID_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("d1", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByID(context.Background(), "d1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestDeptRepo_FindByCode_Found(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "code"}).AddRow("d1", "IT", "IT001")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE code = $1 AND deleted_at IS NULL LIMIT $2`)).
		WithArgs("IT001", 1).
		WillReturnRows(rows)

	dept, err := repo.FindByCode(context.Background(), "IT001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept == nil || dept.Code != "IT001" {
		t.Fatalf("expected dept IT001, got %#v", dept)
	}
}

func TestDeptRepo_FindByCode_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE code = $1 AND deleted_at IS NULL LIMIT $2`)).
		WithArgs("NOPE", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	dept, err := repo.FindByCode(context.Background(), "NOPE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept != nil {
		t.Fatalf("expected nil, got %#v", dept)
	}
}

func TestDeptRepo_FindByCode_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE code = $1 AND deleted_at IS NULL LIMIT $2`)).
		WithArgs("IT", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByCode(context.Background(), "IT")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestDeptRepo_Create_OK(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-dept-id"))
	mock.ExpectCommit()

	dept := &models.Department{Name: "IT", Code: "IT001"}
	created, err := repo.Create(context.Background(), dept)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "new-dept-id" {
		t.Fatalf("expected id new-dept-id, got %q", created.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeptRepo_Create_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("insert error")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "departments"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), &models.Department{Name: "IT", Code: "IT001"})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected insert error, got %v", err)
	}
}

func TestDeptRepo_Update_OK(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "departments"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &models.Department{ID: "d1", Name: "Updated", Code: "UPD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeptRepo_Update_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "departments"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &models.Department{ID: "d1", Name: "X", Code: "X"})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestDeptRepo_List_OK(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "departments"`).WillReturnRows(countRows)

	dataRows := sqlmock.NewRows([]string{"id", "name", "code", "leader_user_id"}).
		AddRow("d1", "IT", "IT001", nil).
		AddRow("d2", "HR", "HR001", nil)
	mock.ExpectQuery(`SELECT \* FROM "departments"`).WillReturnRows(dataRows)

	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	depts, total, err := repo.List(context.Background(), DepartmentFilter{}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(depts) != 2 {
		t.Fatalf("expected 2 depts, got %d", len(depts))
	}
}

func TestDeptRepo_List_CountError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "departments"`).WillReturnError(dbErr)

	_, _, err := repo.List(context.Background(), DepartmentFilter{}, 0, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestDeptRepo_List_SelectError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	dbErr := errors.New("select error")
	mock.ExpectQuery(`SELECT \* FROM "departments"`).WillReturnError(dbErr)

	_, _, err := repo.List(context.Background(), DepartmentFilter{}, 0, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected select error, got %v", err)
	}
}

func TestDeptRepo_SoftDelete_OK(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "departments"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SoftDelete(context.Background(), "d1", "actor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeptRepo_FindByLeaderUserID_Found(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "code", "leader_user_id"}).
		AddRow("d1", "IT", "IT001", "u1")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE leader_user_id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnRows(rows)

	dept, err := repo.FindByLeaderUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept == nil || dept.ID != "d1" {
		t.Fatalf("expected dept d1, got %#v", dept)
	}
}

func TestDeptRepo_FindByLeaderUserID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE leader_user_id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("nobody", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	dept, err := repo.FindByLeaderUserID(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept != nil {
		t.Fatalf("expected nil, got %#v", dept)
	}
}

func TestDeptRepo_FindByLeaderUserID_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE leader_user_id = $1 AND deleted_at IS NULL ORDER BY "departments"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByLeaderUserID(context.Background(), "u1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestDeptRepo_CreateInTx_OK(t *testing.T) {
	_, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	sqlDB2, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB2.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB2, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm: %v", err)
	}
	_ = mock

	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-d"))
	innerMock.ExpectCommit()

	repo := NewDepartmentRepo(innerDB)
	tx := innerDB.Begin()
	if err := repo.CreateInTx(tx, &models.Department{Name: "X", Code: "X1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tx.Commit()
	if err := innerMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeptRepo_List_WithSearch(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "departments"`).WillReturnRows(countRows)

	dataRows := sqlmock.NewRows([]string{"id", "name", "code", "leader_user_id"}).
		AddRow("d1", "IT", "IT001", nil)
	mock.ExpectQuery(`SELECT \* FROM "departments"`).WillReturnRows(dataRows)
	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	depts, total, err := repo.List(context.Background(), DepartmentFilter{Search: "IT"}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(depts) != 1 {
		t.Fatalf("expected 1 dept, got total=%d len=%d", total, len(depts))
	}
}

func TestDeptRepo_SoftDelete_DBError(t *testing.T) {
	repo, mock, cleanup := newMockDeptRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "departments"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.SoftDelete(context.Background(), "d1", "actor")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}
