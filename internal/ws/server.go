package ws

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/service"
)

type Server struct {
	ctx            context.Context
	cancel         context.CancelFunc
	rooms          map[uuid.UUID]*room
	messageService *service.MessageService
	mu             sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	HandshakeTimeout: 10,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		if u.Hostname() == "localhost" {
			return true
		}
		return false
	},
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return &apperror.ChatConnectionError{
			Msg: err.Error(),
		}
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
		go s.runRoom(room, chatId)
	}
	s.mu.Unlock()

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)

	return nil
}

func (s *Server) runRoom(room *room, chatId uuid.UUID) {
	defer func() {
		s.mu.Lock()
		delete(s.rooms, chatId)
		s.mu.Unlock()
		for u := range room.users {
			close(u.send)
		}
		room.cancel()
	}()

	for {
		select {
		case <-room.ctx.Done():
			return

		case user := <-room.addUser:
			room.users[user] = true

		case user := <-room.removeUser:
			if _, ok := room.users[user]; ok {
				delete(room.users, user)
				close(user.send)
			}
			if len(room.users) == 0 {
				return
			}

		case msg := <-room.broadcast:
			go func(m message) {
				_, err := s.messageService.CreateMessage(
					room.ctx,
					m.userID,
					chatId,
					string(m.data),
				)
				if err != nil {
					log.Println("saving message error")
				}
			}(msg)

			for user := range room.users {
				if user.id == msg.userID {
					continue
				}

				select {
				case user.send <- msg:
				default:
					delete(room.users, user)
					close(user.send)
				}
			}
		}
	}
}

func (s *Server) Shutdown() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range s.rooms {
		room.cancel()
	}
}
