package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type            string `json:"type"`
	UserID          string `json:"user_id,omitempty"`
	ApplicationID   string `json:"application_id,omitempty"`
	ApplicationCode string `json:"application_code,omitempty"`
	Status          string `json:"status,omitempty"`
	Message         string `json:"message,omitempty"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Add(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[userID][conn] = struct{}{}
}

func (h *Hub) Remove(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		return
	}

	delete(h.clients[userID], conn)

	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

func (h *Hub) SendToUser(userID string, msg Message) {
	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for conn := range conns {
		_ = conn.WriteJSON(msg)
	}
}

func (h *Hub) Broadcast(msg Message) {
	payload, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.clients {
		for conn := range conns {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}
