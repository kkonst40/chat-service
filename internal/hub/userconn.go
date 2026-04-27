package hub

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kkonst40/chat-service/internal/domain/event"
)

const (
	pongWait           = 60 * time.Second
	pingPeriod         = 45 * time.Second
	writeWait          = 10 * time.Second
	sendChanBufferSize = 64
)

type UserConn struct {
	conn *websocket.Conn
	send chan event.Event
	done chan struct{}
}

func NewUserConn(conn *websocket.Conn) *UserConn {
	return &UserConn{
		conn: conn,
		send: make(chan event.Event, sendChanBufferSize),
		done: make(chan struct{}),
	}
}

func (u *UserConn) writeToConn() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		u.conn.SetWriteDeadline(time.Now().Add(writeWait))
		u.conn.WriteMessage(websocket.CloseMessage, []byte{})
		u.conn.Close()
	}()

	for {
		select {
		case <-u.done:
			return

		case event, ok := <-u.send:
			if !ok {
				return
			}

			u.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := u.conn.WriteJSON(event); err != nil {
				slog.Error("Error writing json to socket")
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

func (u *UserConn) readFromConn() {
	defer func() {
		u.conn.Close()
	}()

	u.conn.SetReadLimit(512)
	u.conn.SetReadDeadline(time.Now().Add(pongWait))

	u.conn.SetPongHandler(func(string) error {
		u.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := u.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
