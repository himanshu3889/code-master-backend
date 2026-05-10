package appWebsocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Constants for buffer and queue managements
const (
	clientBufferSize = 20
)

// Upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for simplicity in this example.
		// In production, restrict this to your domain.
		return true
	},
}

// handles WebSocket requests for connections.
func WsHandler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("Upgrade error")
		return
	}

	conn.SetReadLimit(1024 * 1024) // Add this: reject messages > 1024 KB

	client := &Client{conn: conn, send: make(chan *websocket.PreparedMessage, clientBufferSize)}

	clientMu.Lock()
	userClient = client
	clientMu.Unlock()

	go client.WritePump() // Client's write goroutine
	client.ReadPump()     // Client's read goroutine (blocks)
}
