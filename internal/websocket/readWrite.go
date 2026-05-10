package appWebsocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// Constants for connection management
const (
	pingPeriod   = 25 * time.Second
	pingInterval = 10 * time.Second
	missedPings  = 2

	writeWait = 10 * time.Second

	// Time to wait for a pong response
	pongWait = missedPings*pingPeriod + writeWait
)

// Client represents a single connected WebSocket client.
type Client struct {
	conn *websocket.Conn
	send chan *websocket.PreparedMessage // Channel to send messages to the client; // This can panic if you try to close again
}

// readPump reads messages from the WebSocket connection.
func (client *Client) ReadPump() {
	defer func() {
		// Clean up the client connection when the goroutine exits
		// logrus.Infof("Client disconnected: %s", client.conn.RemoteAddr())
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(pongWait)) // Add deadline
	client.conn.SetPongHandler(func(string) error {       // Add pong handler
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, recMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// logrus.WithError(err).Error("Websocket read error")
			}
			break
		}

		// based on event we will do some more things before broadcast
		client.handleIncomingMessage(recMessage)

	}
}

// writePump writes messages to the WebSocket connection.
func (client *Client) WritePump() {
	ticker := time.NewTicker(pingInterval) // Ping interval
	defer func() {
		// logrus.Infof("Client disconnected by server: %s", client.conn.RemoteAddr())
		ticker.Stop()
		client.conn.Close()
	}()

	client.conn.EnableWriteCompression(true)
	client.conn.SetCompressionLevel(3) // 40% compression, will add CPU overhead but save bandwidth

	for {
		select {
		case preparedMsg, ok := <-client.send:
			if !ok {
				// Hub closed the channel
				client.conn.SetWriteDeadline(time.Now().Add(writeWait))
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WritePreparedMessage(preparedMsg); err != nil {
				// logrus.WithError(err).Error("Write error")
				return
			}
		case <-ticker.C:
			// Send ping messages to keep the connection alive
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// logrus.WithError(err).Error("Websocket ping error")
				return
			}
		}
	}
}
