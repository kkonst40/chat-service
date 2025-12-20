package server

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WsServer struct {
	rooms map[uuid.UUID]*Room
	mu    sync.RWMutex
}

func NewWsServer() *WsServer {
	return &WsServer{
		rooms: make(map[uuid.UUID]*Room),
	}
}

type Message struct {
	userID uuid.UUID
	data   []byte
}

type User struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan Message
}

type Room struct {
	users      map[*User]bool
	addUser    chan *User
	removeUser chan *User
	broadcast  chan Message
	mutex      sync.RWMutex
}

func NewRoom() *Room {
	return &Room{
		users:      make(map[*User]bool),
		addUser:    make(chan *User),
		removeUser: make(chan *User),
		broadcast:  make(chan Message),
	}
}

func (r *Room) Run() {
	for {
		select {
		case user := <-r.addUser:
			r.mutex.Lock()
			r.users[user] = true
			r.mutex.Unlock()

		case user := <-r.removeUser:
			r.mutex.Lock()
			if _, ok := r.users[user]; ok {
				delete(r.users, user)
				close(user.send)
			}
			r.mutex.Unlock()

		case message := <-r.broadcast:
			r.mutex.RLock()
			for user := range r.users {
				if user.id != message.userID {
					select {
					case user.send <- message:
					default:
						delete(r.users, user)
						close(user.send)
					}
				}
			}
			r.mutex.RUnlock()
		}
	}
}

func (u *User) readMessage(r *Room) {
	defer func() {
		r.removeUser <- u
		u.conn.Close()
	}()

	for {
		_, messageBytes, err := u.conn.ReadMessage()
		if err != nil {
			break
		}
		r.broadcast <- Message{
			userID: u.id,
			data:   messageBytes,
		}
	}
}

func (u *User) writeMessage() {
	defer func() {
		u.conn.WriteMessage(websocket.CloseMessage, []byte{})
		u.conn.Close()
	}()

	for message := range u.send {
		err := u.conn.WriteMessage(websocket.TextMessage, message.data)
		if err != nil {
			break
		}
	}
}

func serveWs(room *Room, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	user := &User{
		id:   userID,
		conn: conn,
		send: make(chan Message, 256),
	}

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)
}
