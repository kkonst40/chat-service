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
		return &apperror.ChatConnectionError{
			Msg: err.Error(),
		}
	}

	user := &user{
		id:   userID,
		conn: conn,
		send: make(chan roomEvent, 256),
	}

	s.mu.Lock()
	room, ok := s.rooms[chatID]
	if !ok {
		room = newRoom(s.ctx)
		s.rooms[chatID] = room
		go s.runRoom(room, chatID)
		slog.Info("Room created", "roomID", chatID)
	}
	s.mu.Unlock()

	room.addUser <- user

	go user.writeMessage()
	go user.readMessage(room)

	return nil
}

func (s *Server) runRoom(room *room, chatID uuid.UUID) {
	defer func() {
		s.mu.Lock()
		delete(s.rooms, chatID)
		s.mu.Unlock()
		for u := range room.users {
			close(u.send)
		}
		room.cancel()
		slog.Info("Room stopped", "roomID", chatID)
	}()

	for {
		select {
		case <-room.ctx.Done():
			slog.Debug("Room context done", "roomID", chatID)
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

		case event := <-room.eventQueue:
			event, err := s.handleEvent(event, chatID)
			if err != nil {
				slog.Error("Handling event error", "errors", err.Error())
				continue
			}

			for user := range room.users {
				select {
				case user.send <- event:
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
}

func (s *Server) handleEvent(event roomEvent, chatID uuid.UUID) (roomEvent, error) {
	switch event.Type {
	case ActionCreate:
		msgID, err := uuid.NewV7()
		if err != nil {
			return roomEvent{}, fmt.Errorf("generating msg id error")
		}

		event.MsgID = msgID

		go func() {
			_, err = s.messageService.CreateMessage(
				s.ctx,
				event.MsgID,
				event.UserID,
				chatID,
				event.Text,
			)

			if err != nil {
				slog.Error("Saving message error", "messageID", event.MsgID)
			}
		}()

	case ActionUpdate:
		err := s.messageService.UpdateMessage(
			s.ctx,
			event.MsgID,
			event.Text,
			event.UserID,
		)

		if err != nil {
			return roomEvent{}, err
		}

	case ActionDelete:
		err := s.messageService.DeleteMessage(
			s.ctx,
			event.MsgID,
			event.UserID,
		)

		if err != nil {
			return roomEvent{}, err
		}
	default:
		return roomEvent{}, fmt.Errorf("Unknown action type")
	}

	return event, nil
}
