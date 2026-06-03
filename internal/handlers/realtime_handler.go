package handlers

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/realtime"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

type RealtimeHandler struct {
	hub *realtime.Hub
}

func NewRealtimeHandler(hub *realtime.Hub) *RealtimeHandler {
	return &RealtimeHandler{
		hub: hub,
	}
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *RealtimeHandler) CitizenNotifications(c *echo.Context) error {
	claims := configs.ClaimsFromContext(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "citizen.unauthorized")
	}

	conn, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	h.hub.Add(claims.ID, conn)

	defer func() {
		h.hub.Remove(claims.ID, conn)
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	return nil
}
