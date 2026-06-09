package repositories

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockReminderLogRepo(t *testing.T) (ApplicationReminderLogRepository, sqlmock.Sqlmock, func()) {
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
	return NewApplicationReminderLogRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestApplicationReminderLogRepoExists(t *testing.T) {
	repo, mock, cleanup := newMockReminderLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "application_reminder_logs" WHERE application_id = $1 AND recipient_user_id = $2 AND reminder_type = $3 AND deleted_at IS NULL`)).
		WithArgs("app-1", "user-1", "due_in_48h").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.Exists("app-1", "user-1", "due_in_48h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected reminder log to exist")
	}
}

func TestApplicationReminderLogRepoCreate(t *testing.T) {
	repo, mock, cleanup := newMockReminderLogRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "application_reminder_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-1"))
	mock.ExpectCommit()

	err := repo.Create(&models.ApplicationReminderLog{
		ApplicationID:   "app-1",
		RecipientUserID: "user-1",
		ReminderType:    "due_in_48h",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
