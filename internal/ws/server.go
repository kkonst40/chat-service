package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kkonst40/ichat/internal/service"
)

type Server struct {
	rooms          map[uuid.UUID]*room
	messageService *service.MessageService
	mu             sync.Mutex
	//chatService    *service.ChatService
}

func NewWsServer(chatService *service.ChatService, messageService *service.MessageService) *Server {
	return &Server{
		rooms:          make(map[uuid.UUID]*room),
		messageService: messageService,
		//chatService:    chatService,
	}
}

func (s *Server) Connect(w http.ResponseWriter, r *http.Request, userId uuid.UUID, chatId uuid.UUID) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	user := &user{
		id:   userId,
		conn: conn,
		send: make(chan message, 256),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[chatId]
	if !ok {
		room = newRoom()
		s.rooms[chatId] = room
		go s.runRoom(chatId)
	}

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)
}

func (s *Server) runRoom(chatId uuid.UUID) error {
	room, ok := s.rooms[chatId]
	if !ok {
		return fmt.Errorf("chat %v does not exist", chatId)
	}

	for {
		select {
		case user := <-room.addUser:
			room.mutex.Lock()
			room.users[user] = true
			room.mutex.Unlock()

		case user := <-room.removeUser:
			room.mutex.Lock()
			if _, ok := room.users[user]; ok {
				delete(room.users, user)
				close(user.send)
			}
			if len(room.users) == 0 {
				close(room.addUser)
				close(room.removeUser)
				close(room.broadcast)
				delete(s.rooms, chatId)
				return nil
			}
			room.mutex.Unlock()

		case message := <-room.broadcast:
			_, err := s.messageService.CreateMessage(message.userID, chatId, string(message.data))
			if err != nil {
				return fmt.Errorf("error saving message error")
			}
			room.mutex.RLock()
			for user := range room.users {
				if user.id != message.userID {
					select {
					case user.send <- message:
					default:
						delete(room.users, user)
						close(user.send)
					}
				}
			}
			room.mutex.RUnlock()
		}
	}
}
