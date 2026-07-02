package websocket

import (
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

// Message WebSocket 推送消息结构
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Time    time.Time   `json:"time"`
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     zerolog.Logger
}

var (
	hubInstance *Hub
	hubOnce     sync.Once
)

// GetHub 获取 Hub 单例
func GetHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			clients:    make(map[*Client]bool),
			broadcast:  make(chan Message, 256),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			logger:     config.GetServiceLogger("websocket"),
		}
		go hubInstance.run()
	})
	return hubInstance
}

// run Hub 主循环
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息到所有连接的客户端
func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

// BroadcastJSON 将任意数据序列化后广播
func (h *Hub) BroadcastJSON(msgType string, payload interface{}) {
	msg := Message{Type: msgType, Payload: payload, Time: time.Now()}
	h.broadcast <- msg
}

// ClientCount 返回当前连接数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
