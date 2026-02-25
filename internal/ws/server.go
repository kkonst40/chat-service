package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	errs "github.com/kkonst40/ichat/internal/errors"
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
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

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

func NewServer(chatService *service.ChatService, messageService *service.MessageService) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:            ctx,
		cancel:         cancel,
		rooms:          make(map[uuid.UUID]*room),
		messageService: messageService,
	}
}

func (s *Server) Connect(w http.ResponseWriter, r *http.Request, userID uuid.UUID, chatID uuid.UUID) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrChatConnection, err)
	}

	user := &user{
		id:   userID,
		conn: conn,
		send: make(chan roomEvent, 256),
	}

	s.mu.Lock()
	room, ok := s.rooms[chatID]
	if !ok {
		room = newRoom(s.ctx, s.messageService)
		s.rooms[chatID] = room
		go func() {
			room.run(chatID)

			s.mu.Lock()
			delete(s.rooms, chatID)
			s.mu.Unlock()
			for u := range room.users {
				close(u.send)
			}
			room.cancel()
			slog.Debug("Room stopped", "roomID", chatID)
		}()

		slog.Debug("Room created", "roomID", chatID)
	}
	s.mu.Unlock()

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)

	return nil
}

func (s *Server) Shutdown() {
	s.cancel()
}
