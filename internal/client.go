package internal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingPeriod = 30 * time.Second
	pongWait   = 60 * time.Second
)

type Client struct {
	id     string
	room   string
	hub    *Hub
	conn   *websocket.Conn
	send   chan Envelope
	active bool
}

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		slog.Error("не удалось сгенерировать id клиента", "error", err)
	}
	return hex.EncodeToString(b)
}

func NewClient(url string, hub *Hub, conn *websocket.Conn, send chan Envelope) *Client {
	client := Client{
		id:     generateID(),
		room:   url,
		conn:   conn,
		send:   send,
		hub:    hub,
		active: true,
	}

	return &client
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case x, ok := <-c.send:
			if !ok {
				return
			}
			message, err := json.Marshal(x)

			if err != nil {
				slog.Error("Была получена ошибка: ", "Error", err)
				continue
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (c *Client) ReadPump() {
	c.hub.register <- c

	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, messageByte, errR := c.conn.ReadMessage()

		if errR != nil || messageType == websocket.CloseMessage {
			if websocket.IsUnexpectedCloseError(errR, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Error("неожиданный разрыв соединения", "error", errR)
			} else {
				slog.Info("клиент отключился", "error", errR)
			}

			c.hub.unregister <- c

			break
		}

		var message Envelope
		errJ := json.Unmarshal(messageByte, &message)

		if errJ != nil {
			slog.Error("Была получена ошибка: ", "Error", errJ)

			continue
		}

		c.hub.inbound <- inbound{
			client: c,
			msg:    message,
		}
	}
}
