package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type NotificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(notif *models.Notification) error {
	return s.repo.Create(notif)
}

func (s *NotificationService) List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error) {
	items, total, err := s.repo.ListByUserID(userID, filter, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dtos.NotificationResponse, 0, len(items))
	for _, n := range items {
		resp = append(resp, dtos.NewNotificationResponse(&n))
	}
	return resp, total, nil
}

func (s *NotificationService) MarkAsRead(id, userID string) error {
	return s.repo.MarkAsRead(id, userID)
}

func (s *NotificationService) MarkAllAsRead(userID string) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *NotificationService) CountUnread(userID string) (int64, error) {
	return s.repo.CountUnread(userID)
}
