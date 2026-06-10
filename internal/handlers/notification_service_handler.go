package handlers

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type notificationAPIService interface {
	List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error)
	CountUnread(userID string) (int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
}

type NotificationServiceHandler struct {
	svc notificationAPIService
}

func NewNotificationServiceHandler(svc notificationAPIService) *NotificationServiceHandler {
	return &NotificationServiceHandler{svc: svc}
}

func (h *NotificationServiceHandler) List(c *echo.Context) error {
	userID := c.QueryParam("user_id")
	page, limit := parsePagination(c)

	filter := repositories.NotificationFilter{
		Type: c.QueryParam("type"),
	}
	if c.QueryParam("is_read") == "true" {
		v := true
		filter.IsRead = &v
	} else if c.QueryParam("is_read") == "false" {
		v := false
		filter.IsRead = &v
	}

	items, total, err := h.svc.List(userID, filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"notifications": items,
		"pagination":    utils.NewPagination(page, limit, total),
	})
}

func (h *NotificationServiceHandler) CountUnread(c *echo.Context) error {
	userID := c.QueryParam("user_id")
	count, err := h.svc.CountUnread(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.JSON(http.StatusOK, utils.Map{"count": count})
}

func (h *NotificationServiceHandler) MarkAsRead(c *echo.Context) error {
	userID := c.QueryParam("user_id")
	if err := h.svc.MarkAsRead(c.Param("id"), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.NoContent(http.StatusOK)
}

func (h *NotificationServiceHandler) MarkAllAsRead(c *echo.Context) error {
	userID := c.QueryParam("user_id")
	if err := h.svc.MarkAllAsRead(userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.NoContent(http.StatusOK)
}
