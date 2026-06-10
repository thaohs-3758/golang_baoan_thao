package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type NotificationAPIService struct {
	*NotificationService
}

func NewNotificationAPIService(repo repositories.NotificationRepository) *NotificationAPIService {
	return &NotificationAPIService{
		NotificationService: NewNotificationService(repo),
	}
}

func (s *NotificationAPIService) List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error) {
	return s.NotificationService.List(userID, filter, page, limit)
}

func (s *NotificationAPIService) CountUnread(userID string) (int64, error) {
	return s.NotificationService.CountUnread(userID)
}

func (s *NotificationAPIService) MarkAsRead(id, userID string) error {
	return s.NotificationService.MarkAsRead(id, userID)
}

func (s *NotificationAPIService) MarkAllAsRead(userID string) error {
	return s.NotificationService.MarkAllAsRead(userID)
}
