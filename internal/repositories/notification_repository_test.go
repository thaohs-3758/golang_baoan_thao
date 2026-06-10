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

func newMockNotificationRepo(t *testing.T) (NotificationRepository, sqlmock.Sqlmock, func()) {
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
	return NewNotificationRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestNotificationRepo_ListByUserID_NoFilter(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "message", "type", "is_read", "created_at"}).
		AddRow("notif-1", "user-1", "Title 1", "Message 1", "system", false, time.Now()).
		AddRow("notif-2", "user-1", "Title 2", "Message 2", "result", true, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notifications" WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2`)).
		WithArgs("user-1", 10).
		WillReturnRows(rows)

	items, total, err := repo.ListByUserID("user-1", NotificationFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_ListByUserID_WithTypeFilter(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE (user_id = $1 AND deleted_at IS NULL) AND type = $2`)).
		WithArgs("user-1", "system").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "message", "type", "is_read", "created_at"}).
		AddRow("notif-1", "user-1", "Title", "Message", "system", false, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notifications" WHERE (user_id = $1 AND deleted_at IS NULL) AND type = $2 ORDER BY created_at DESC LIMIT $3`)).
		WithArgs("user-1", "system", 10).
		WillReturnRows(rows)

	items, total, err := repo.ListByUserID("user-1", NotificationFilter{Type: "system"}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].ID != "notif-1" {
		t.Fatalf("unexpected items: %v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_ListByUserID_WithIsReadFilter(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	isRead := false
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE (user_id = $1 AND deleted_at IS NULL) AND is_read = $2`)).
		WithArgs("user-1", false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "message", "type", "is_read", "created_at"}).
		AddRow("notif-1", "user-1", "Unread", "Msg", "system", false, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notifications" WHERE (user_id = $1 AND deleted_at IS NULL) AND is_read = $2 ORDER BY created_at DESC LIMIT $3`)).
		WithArgs("user-1", false, 10).
		WillReturnRows(rows)

	items, total, err := repo.ListByUserID("user-1", NotificationFilter{IsRead: &isRead}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got total=%d, items=%d", total, len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_ListByUserID_CountError(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("user-1").
		WillReturnError(dbErr)

	_, _, err := repo.ListByUserID("user-1", NotificationFilter{}, 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_MarkAsRead(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "notifications" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "notif-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkAsRead("notif-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_MarkAsRead_Error(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	dbErr := errors.New("update error")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "notifications" SET`)).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.MarkAsRead("notif-1", "user-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_MarkAllAsRead(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "notifications" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "user-1").
		WillReturnResult(sqlmock.NewResult(3, 3))
	mock.ExpectCommit()

	err := repo.MarkAllAsRead("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_Create(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	notif := &models.Notification{
		ID:        "notif-new",
		UserID:    "user-1",
		Title:     "Hello",
		Message:   "World",
		Type:      models.NotificationTypeSystem,
		CreatedAt: time.Now(),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "notifications"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("notif-new"))
	mock.ExpectCommit()

	err := repo.Create(notif)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_Create_SetsCreatedAt(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	notif := &models.Notification{
		UserID:  "user-1",
		Title:   "Auto time",
		Message: "msg",
		Type:    models.NotificationTypeSystem,
		// CreatedAt intentionally zero
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "notifications"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("notif-auto"))
	mock.ExpectCommit()

	err := repo.Create(notif)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notif.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_Create_DuplicateSourceEventIDIsIgnored(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	notif := &models.Notification{
		UserID:  "user-1",
		Title:   "Hello",
		Message: "World",
		Type:    models.NotificationTypeReceived,
	}
	eventID := "evt-submit-1"
	notif.SourceEventID = &eventID

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "notifications"`)).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	err := repo.Create(notif)
	if err != nil {
		t.Fatalf("expected nil on duplicate source event id, got %v", err)
	}
}

func TestNotificationRepo_CountUnread(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE user_id = $1 AND is_read = false AND deleted_at IS NULL`)).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.CountUnread("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected count 5, got %d", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNotificationRepo_CountUnread_Error(t *testing.T) {
	repo, mock, cleanup := newMockNotificationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "notifications" WHERE user_id = $1 AND is_read = false AND deleted_at IS NULL`)).
		WithArgs("user-1").
		WillReturnError(dbErr)

	_, err := repo.CountUnread("user-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
