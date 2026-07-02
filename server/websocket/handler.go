package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 认证已在 Gin 中间件完成
	},
}

// WSHandler 将 HTTP 连接升级为 WebSocket
func WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	client := &Client{
		hub:    GetHub(),
		conn:   conn,
		send:   make(chan Message, 64),
		userID: userID.(string),
		role:   role.(string),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
