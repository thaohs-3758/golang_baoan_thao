package repositories

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type NotificationFilter struct {
	Type   string
	IsRead *bool
}

type NotificationRepository interface {
	ListByUserID(userID string, filter NotificationFilter, page, limit int) ([]models.Notification, int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	Create(notif *models.Notification) error
	CountUnread(userID string) (int64, error)
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) ListByUserID(userID string, filter NotificationFilter, page, limit int) ([]models.Notification, int64, error) {
	q := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.IsRead != nil {
		q = q.Where("is_read = ?", *filter.IsRead)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Notification
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *notificationRepo) MarkAsRead(id, userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *notificationRepo) MarkAllAsRead(userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false AND deleted_at IS NULL", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *notificationRepo) Create(notif *models.Notification) error {
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = time.Now()
	}
	err := r.db.Create(notif).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) && notif.SourceEventID != nil && *notif.SourceEventID != "" {
		return nil
	}
	return err
}

func (r *notificationRepo) CountUnread(userID string) (int64, error) {
	var n int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}
