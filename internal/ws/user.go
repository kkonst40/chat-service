package ws

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 55 * time.Second
)

type user struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan message
}

func (u *user) readMessage(r *room) {
	defer func() {
		r.removeUser <- u
		u.conn.Close()
	}()

	u.conn.SetReadLimit(512 * 1024)
	u.conn.SetReadDeadline(time.Now().Add(pongWait))
	u.conn.SetPongHandler(func(string) error {
		u.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := u.conn.ReadMessage()
		if err != nil {
			break
		}
		r.broadcast <- message{
			userID: u.id,
			data:   messageBytes,
		}
	}
}

func (u *user) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		u.conn.WriteMessage(websocket.CloseMessage, []byte{})
		u.conn.Close()
	}()

	for {
		select {
		case message, ok := <-u.send:
			if !ok {
				return
			}
			msg := jsonMessage{
				UserID: message.userID.String(),
				Text:   string(message.data),
			}
			if err := u.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			u.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := u.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
