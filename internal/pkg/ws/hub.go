package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
)

// Message WebSocket 消息
type Message struct {
	Type string      `json:"type"` // task_progress / task_result / task_done
	Data interface{} `json:"data"`
}

// Hub WebSocket 连接管理
type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*websocket.Conn]bool // room -> connections
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

// Join 加入房间
func (h *Hub) Join(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*websocket.Conn]bool)
	}
	h.rooms[room][conn] = true
}

// Leave 离开房间
func (h *Hub) Leave(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.rooms[room]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.rooms, room)
		}
	}
}

// Broadcast 向房间广播消息
func (h *Hub) Broadcast(room string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Errorf("ws marshal: %v", err)
		return
	}

	h.mu.RLock()
	conns := h.rooms[room]
	h.mu.RUnlock()

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Log.Debugf("ws write: %v", err)
			h.Leave(room, conn)
			conn.Close()
		}
	}
}
