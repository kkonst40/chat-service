package ws

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 45 * time.Second
	writeWait  = 10 * time.Second
)

type user struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan roomEvent
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
		var receiveEvent roomEvent
		err := u.conn.ReadJSON(&receiveEvent)
		if err != nil {
			fmt.Println("error reading json from socket")
			break
		}

		u.conn.SetReadDeadline(time.Now().Add(pongWait))

		switch receiveEvent.Type {
		case ActionCreate, ActionUpdate, ActionDelete:
		default:
			continue
		}

		receiveEvent.UserID = u.id

		r.eventQueue <- receiveEvent
	}
}

func (u *user) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		u.conn.SetWriteDeadline(time.Now().Add(writeWait))
		u.conn.WriteMessage(websocket.CloseMessage, []byte{})
		u.conn.Close()
	}()

	for {
		select {
		case event, ok := <-u.send:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				fmt.Println("user send chan is closed")
				return
			}

			if err := u.conn.WriteJSON(event); err != nil {
				fmt.Println("error writing json to socket")
				return
			}

		case <-ticker.C:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := u.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
