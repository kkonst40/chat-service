package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kkonst40/ichat/internal/service"
)

type Server struct {
	ctx            context.Context
	cancel         context.CancelFunc
	rooms          map[uuid.UUID]*room
	messageService *service.MessageService
	mu             sync.Mutex
}

func NewWsServer(chatService *service.ChatService, messageService *service.MessageService) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:            ctx,
		cancel:         cancel,
		rooms:          make(map[uuid.UUID]*room),
		messageService: messageService,
	}
}

func (s *Server) Connect(w http.ResponseWriter, r *http.Request, userId uuid.UUID, chatId uuid.UUID) error {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	user := &user{
		id:   userId,
		conn: conn,
		send: make(chan message, 256),
	}

	s.mu.Lock()

	room, ok := s.rooms[chatId]
	if !ok {
		room = newRoom(s.ctx)
		s.rooms[chatId] = room
		go s.runRoom(chatId)
	}
	s.mu.Unlock()

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)
	return nil
}

func (s *Server) runRoom(chatId uuid.UUID) error {
	room, ok := s.rooms[chatId]
	if !ok {
		return fmt.Errorf("chat %v does not exist", chatId)
	}

	for {
		select {
		case <-room.ctx.Done():
			return room.ctx.Err()

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
				room.cancel()
				delete(s.rooms, chatId)
				return nil
			}
			room.mutex.Unlock()

		case message := <-room.broadcast:
			_, err := s.messageService.CreateMessage(
				room.ctx,
				message.userID,
				chatId,
				string(message.data),
			)
			if err != nil {
				log.Println("error saving message error")
				continue
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
