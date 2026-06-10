package repositories

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockApplicationRepo(t *testing.T) (ApplicationRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}
	return NewApplicationRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestApplicationRepoListByCitizenReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id"}).
		AddRow("app1", "APP-20240101-ABCDEF", "u1")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	// Preload ServiceType
	stRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	items, total, err := repo.ListByCitizen("u1", 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestApplicationRepoListByCitizenCountError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnError(dbErr)

	_, _, err := repo.ListByCitizen("u1", 1, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestApplicationRepoListByCitizenFindError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnRows(countRows)

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, _, err := repo.ListByCitizen("u1", 1, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestApplicationRepoGetByIDForCitizenReturnsApp(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "citizen_user_id"}).
		AddRow("app1", "u1")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	stRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	app, err := repo.GetByIDForCitizen("app1", "u1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if app == nil || app.ID != "app1" {
		t.Fatalf("expected app1, got %#v", app)
	}
}

func TestApplicationRepoGetByIDForCitizenNotFound(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.GetByIDForCitizen("bad-id", "u1")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestApplicationRepoGetByIDPreloadsCitizenUser(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id", "service_type_id", "assigned_staff_user_id"}).
		AddRow("app1", "APP-1", "citizen-1", "service-1", nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "applications" WHERE id = $1 AND deleted_at IS NULL ORDER BY "applications"."id" LIMIT $2`)).
		WithArgs("app1", 1).
		WillReturnRows(appRows)

	attachRows := sqlmock.NewRows([]string{"id", "application_id"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_attachments" WHERE "application_attachments"."application_id" = $1`)).
		WithArgs("app1").
		WillReturnRows(attachRows)

	serviceRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("service-1", "Service A")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "service_types" WHERE "service_types"."id" = $1`)).
		WithArgs("service-1").
		WillReturnRows(serviceRows)

	citizenRows := sqlmock.NewRows([]string{"id", "name", "email"}).AddRow("citizen-1", "Citizen A", "citizen@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs("citizen-1").
		WillReturnRows(citizenRows)

	app, err := repo.GetByID("app1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if app == nil {
		t.Fatal("expected app, got nil")
	}
	if app.CitizenUser.Name != "Citizen A" {
		t.Fatalf("expected citizen name %q, got %q", "Citizen A", app.CitizenUser.Name)
	}
}

func TestApplicationRepoListStatusLogsByCitizenReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnRows(countRows)

	logRows := sqlmock.NewRows([]string{"id", "application_id", "new_status", "created_at"}).
		AddRow("log1", "app1", "processing", time.Now())
	mock.ExpectQuery(`SELECT`).WillReturnRows(logRows)

	userRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("staff1", "Staff A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(userRows)

	items, total, err := repo.ListStatusLogsByCitizen("app1", "u1", 1, 10, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestApplicationRepoListStatusLogsByCitizenWithSince(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	since := time.Now().Add(-24 * time.Hour)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnRows(countRows)

	logRows := sqlmock.NewRows([]string{"id", "application_id", "new_status", "created_at"}).
		AddRow("log1", "app1", "processing", time.Now())
	mock.ExpectQuery(`SELECT`).WillReturnRows(logRows)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, total, err := repo.ListStatusLogsByCitizen("app1", "u1", 1, 10, &since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got total=%d len=%d", total, len(items))
	}
}

func TestApplicationRepoListStatusLogsByCitizenCountError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnError(dbErr)

	_, _, err := repo.ListStatusLogsByCitizen("app1", "u1", 1, 10, nil)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestApplicationRepoListStatusLogsByCitizenFindError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, _, err := repo.ListStatusLogsByCitizen("app1", "u1", 1, 10, nil)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestApplicationRepoAdminListReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id", "service_type_id", "status"}).
		AddRow("app1", "APP-1", "c1", "s1", "received").
		AddRow("app2", "APP-2", "c2", "s2", "processing")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	// Preloads for ServiceType, CitizenUser, AssignedStaffUser
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	items, total, err := repo.AdminList(ApplicationFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestApplicationRepoAdminListWithFilters(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "status"}).AddRow("app1", "received")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, _, err := repo.AdminList(ApplicationFilter{Status: "received", Service: "CCCD", Submitter: "An"}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplicationRepoAdminListCountError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnError(dbErr)

	_, _, err := repo.AdminList(ApplicationFilter{}, 1, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestApplicationRepoUpdateAssignedStaff(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	staffID := "staff-1"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "applications"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateAssignedStaff("app-1", &staffID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationRepoUpdateAssignedStaff_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("update error")
	staffID := "staff-1"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "applications"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.UpdateAssignedStaff("app-1", &staffID, "admin-1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestApplicationRepoGetDashboardStats(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"status", "count"}).
		AddRow("received", 5).
		AddRow("processing", 3).
		AddRow("approved", 2).
		AddRow("rejected", 1)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	stats, err := repo.GetDashboardStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 11 {
		t.Fatalf("expected total=11, got %d", stats.Total)
	}
	if stats.Approved != 2 {
		t.Fatalf("expected approved=2, got %d", stats.Approved)
	}
	if stats.Rejected != 1 {
		t.Fatalf("expected rejected=1, got %d", stats.Rejected)
	}
	if stats.Pending != 8 {
		t.Fatalf("expected pending=8, got %d", stats.Pending)
	}
}

func TestApplicationRepoGetDashboardStats_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("scan error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.GetDashboardStats()
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected scan error, got %v", err)
	}
}

func TestApplicationRepoListRecent(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "application_code"}).
		AddRow("app1", "APP-1").AddRow("app2", "APP-2")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	items, err := repo.ListRecent(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestApplicationRepoListRecent_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.ListRecent(5)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestApplicationRepoGetDashboardStatsForStaff(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"status", "count"}).
		AddRow("processing", 4).
		AddRow("approved", 3).
		AddRow("rejected", 2).
		AddRow("need_more_info", 1)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	stats, err := repo.GetDashboardStatsForStaff("staff-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 10 {
		t.Fatalf("expected total=10, got %d", stats.Total)
	}
	if stats.Approved != 3 {
		t.Fatalf("expected approved=3, got %d", stats.Approved)
	}
	if stats.Rejected != 2 {
		t.Fatalf("expected rejected=2, got %d", stats.Rejected)
	}
}

func TestApplicationRepoGetDashboardStatsForStaff_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("scan error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.GetDashboardStatsForStaff("staff-1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected scan error, got %v", err)
	}
}

func TestApplicationRepoListRecentForStaff(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "application_code"}).AddRow("app1", "APP-1")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	items, err := repo.ListRecentForStaff("staff-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestApplicationRepoListRecentForStaff_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.ListRecentForStaff("staff-1", 5)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestApplicationRepoProcessStatusUpdate(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	oldStatus := models.ApplicationStatusReceived
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "applications"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO "application_status_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-1"))
	mock.ExpectCommit()

	err := repo.ProcessStatusUpdate("app-1", &oldStatus, models.ApplicationStatusProcessing, "", "", nil, nil, "staff-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationRepoProcessStatusUpdate_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	oldStatus := models.ApplicationStatusReceived
	dbErr := errors.New("update error")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "applications"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.ProcessStatusUpdate("app-1", &oldStatus, models.ApplicationStatusProcessing, "", "", nil, nil, "staff-1", nil)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestApplicationRepoIsApplicationCodeConflict(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{gorm.ErrDuplicatedKey, true},
		{errors.New("pq: duplicate key value violates unique constraint \"applications_application_code\""), true},
		{errors.New("some other error"), false},
	}
	for _, tc := range cases {
		got := isApplicationCodeConflict(tc.err)
		if got != tc.want {
			t.Errorf("isApplicationCodeConflict(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestApplicationRepoCreateWithAttachments_Success(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	app := &models.Application{ApplicationCode: "APP-1"}
	atts := []models.ApplicationAttachment{{FileName: "doc.pdf"}}

	mock.ExpectBegin()
	// Insert application
	mock.ExpectQuery(`INSERT INTO "applications"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("app-1"))
	// Insert attachments
	mock.ExpectQuery(`INSERT INTO "application_attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("att-1"))
	mock.ExpectCommit()

	err := repo.CreateWithAttachments(app, atts, func() string { return "APP-2" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationRepoCreateWithAttachments_Success_WithoutNotificationInsert(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	app := &models.Application{ApplicationCode: "APP-1"}
	atts := []models.ApplicationAttachment{{FileName: "doc.pdf"}}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "applications"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("app-1"))
	mock.ExpectQuery(`INSERT INTO "application_attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("att-1"))
	mock.ExpectCommit()

	err := repo.CreateWithAttachments(app, atts, func() string { return "APP-2" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationRepoCreateWithAttachments_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	app := &models.Application{ApplicationCode: "APP-1"}

	dbErr := errors.New("unique constraint")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "applications"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.CreateWithAttachments(app, nil, func() string { return "APP-2" })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApplicationRepoCreateAttachmentsEmpty(t *testing.T) {
	repo, _, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	err := repo.CreateAttachments("app1", []models.ApplicationAttachment{})
	if err != nil {
		t.Fatalf("expected nil for empty attachments, got %v", err)
	}
}

func TestApplicationRepoGetByID_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.GetByID("bad-id")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestApplicationRepoProcessStatusUpdate_WithAttachments(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	oldStatus := models.ApplicationStatusReceived
	atts := []models.ApplicationAttachment{{FileName: "result.pdf", AttachmentType: models.AttachmentTypeResult}}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "applications"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO "application_attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("att-1"))
	mock.ExpectQuery(`INSERT INTO "application_status_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-1"))
	mock.ExpectCommit()

	err := repo.ProcessStatusUpdate("app-1", &oldStatus, models.ApplicationStatusProcessing, "", "", nil, nil, "staff-1", atts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplicationRepoCreateWithAttachments_ConflictRetry(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	app := &models.Application{ApplicationCode: "APP-1"}

	conflictErr := errors.New(`pq: duplicate key value violates unique constraint "applications_application_code"`)

	// First attempt: conflict on insert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "applications"`).WillReturnError(conflictErr)
	mock.ExpectRollback()

	// Second attempt: success
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "applications"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("app-2"))
	mock.ExpectCommit()

	err := repo.CreateWithAttachments(app, nil, func() string { return "APP-2" })
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
}

func TestApplicationRepoCreateAttachmentsSuccess(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "application_attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("att1"))
	mock.ExpectCommit()

	err := repo.CreateAttachments("app1", []models.ApplicationAttachment{{FileName: "a.pdf", AttachmentType: models.AttachmentTypeSupplement}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestApplicationRepoAdminList_WithAssignedStaffFilter(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\).*assigned_staff_user_id`).
		WithArgs("staff-1").
		WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id", "service_type_id", "assigned_staff_user_id"}).
		AddRow("app1", "APP-1", "citizen-1", "service-1", "staff-1")
	mock.ExpectQuery(`SELECT .*assigned_staff_user_id`).
		WithArgs("staff-1", 10, 0).
		WillReturnRows(appRows)

	serviceRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("service-1", "Service A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(serviceRows)
	citizenRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("citizen-1", "Citizen A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(citizenRows)
	staffRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("staff-1", "Staff A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(staffRows)

	items, total, err := repo.AdminList(ApplicationFilter{AssignedStaffUserID: "staff-1"}, 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestApplicationRepoListDueWithin(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "application_code", "service_type_id", "citizen_user_id", "status",
		"submitted_data", "submitted_at", "created_at", "updated_at", "due_at", "assigned_staff_user_id",
	}).AddRow("app-1", "HS001", "svc-1", "citizen-1", "processing", []byte(`{}`), now, now, now, now.Add(24*time.Hour), "staff-1")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "applications" WHERE deleted_at IS NULL AND due_at IS NOT NULL AND (due_at > $1 AND due_at <= $2) AND status IN ($3,$4,$5)`)).
		WithArgs(now, now.Add(48*time.Hour), models.ApplicationStatusReceived, models.ApplicationStatusProcessing, models.ApplicationStatusNeedMoreInfo).
		WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "service_types" WHERE "service_types"."id" = $1 AND deleted_at IS NULL`)).
		WithArgs("svc-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "responsible_department_id"}).AddRow("svc-1", "Service A", "dept-1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "departments" WHERE "departments"."id" = $1 AND deleted_at IS NULL`)).
		WithArgs("dept-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "leader_user_id"}).AddRow("dept-1", "Dept A", "manager-1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND deleted_at IS NULL`)).
		WithArgs("manager-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("manager-1", "Manager A"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND deleted_at IS NULL`)).
		WithArgs("staff-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("staff-1", "Staff A"))

	items, err := repo.ListDueWithin(now, now.Add(48*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
