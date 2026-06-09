package jobs

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
)

func TestTemporaryFileCleanupJobDeletesOnlyExpiredOrphans(t *testing.T) {
	now := time.Now()
	fs := &fakeTempStorage{
		dirs: []utils.TempDirInfo{
			{ApplicationID: "orphan-1", CreatedAt: now.Add(-25 * time.Hour)},
			{ApplicationID: "active-1", CreatedAt: now.Add(-25 * time.Hour)},
			{ApplicationID: "fresh-1", CreatedAt: now.Add(-2 * time.Hour)},
		},
	}

	repo := &fakeCleanupApplicationRepo{
		owned: map[string]bool{
			"active-1": true,
		},
	}

	job := NewTemporaryFileCleanupJob(fs, repo, 24*time.Hour, func() time.Time { return now })
	result, err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Deleted != 1 {
		t.Fatalf("expected 1 deletion, got %d", result.Deleted)
	}
	if len(fs.deleted) != 1 || fs.deleted[0] != "orphan-1" {
		t.Fatalf("expected orphan-1 to be deleted, got %#v", fs.deleted)
	}
}

type fakeTempStorage struct {
	dirs    []utils.TempDirInfo
	deleted []string
	err     error
}

func (s *fakeTempStorage) SaveApplicationFile(string, *multipart.FileHeader) (string, string, int64, error) {
	return "", "", 0, nil
}

func (s *fakeTempStorage) RemoveApplicationDir(applicationID string) error {
	s.deleted = append(s.deleted, applicationID)
	return s.err
}

func (s *fakeTempStorage) RemoveFile(string) error { return nil }

func (s *fakeTempStorage) ListApplicationDirs() ([]utils.TempDirInfo, error) {
	return s.dirs, s.err
}

type fakeCleanupApplicationRepo struct {
	owned map[string]bool
	err   error
}

func (r *fakeCleanupApplicationRepo) AttachmentExistsForApplication(applicationID string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return r.owned[applicationID], nil
}
