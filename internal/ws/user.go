package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	defer func() {
		u.conn.WriteMessage(websocket.CloseMessage, []byte{})
		u.conn.Close()
	}()

	for message := range u.send {
		msg := jsonMessage{
			UserID: message.userID.String(),
			Text:   string(message.data),
		}
		err := u.conn.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}
