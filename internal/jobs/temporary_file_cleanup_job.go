package jobs

import (
	"context"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
)

type tempStorage interface {
	ListApplicationDirs() ([]utils.TempDirInfo, error)
	RemoveApplicationDir(applicationID string) error
}

type cleanupApplicationRepo interface {
	AttachmentExistsForApplication(applicationID string) (bool, error)
}

type TemporaryFileCleanupJob struct {
	storage tempStorage
	appRepo cleanupApplicationRepo
	ttl     time.Duration
	now     func() time.Time
}

func NewTemporaryFileCleanupJob(storage tempStorage, appRepo cleanupApplicationRepo, ttl time.Duration, now func() time.Time) *TemporaryFileCleanupJob {
	if now == nil {
		now = time.Now
	}
	return &TemporaryFileCleanupJob{storage: storage, appRepo: appRepo, ttl: ttl, now: now}
}

func (j *TemporaryFileCleanupJob) Run(ctx context.Context) (Result, error) {
	_ = ctx
	startedAt := j.now()
	result := Result{JobName: "temporary_file_cleanup", StartedAt: startedAt}

	dirs, err := j.storage.ListApplicationDirs()
	if err != nil {
		result.Errors++
		result.FinishedAt = j.now()
		return result, err
	}
	result.Scanned = len(dirs)

	for _, dir := range dirs {
		if startedAt.Sub(dir.CreatedAt) < j.ttl {
			result.Skipped++
			continue
		}

		owned, err := j.appRepo.AttachmentExistsForApplication(dir.ApplicationID)
		if err != nil {
			result.Errors++
			continue
		}
		if owned {
			result.Skipped++
			continue
		}

		if err := j.storage.RemoveApplicationDir(dir.ApplicationID); err != nil {
			result.Errors++
			continue
		}
		result.Deleted++
	}

	result.FinishedAt = j.now()
	return result, nil
}
