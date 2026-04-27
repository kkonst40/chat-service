package hub

import (
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/event"
)

type Hub struct {
	users map[uuid.UUID]map[*UserConn]struct{}
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		users: make(map[uuid.UUID]map[*UserConn]struct{}),
	}
}

func (h *Hub) ServeConn(userID uuid.UUID, u *UserConn) {
	defer func() {
		h.Unregister(userID, u)
		close(u.send)
		u.conn.Close()
	}()

	h.Register(userID, u)

	go u.writeToConn()
	u.readFromConn()
}

func (h *Hub) Register(userID uuid.UUID, conn *UserConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.users[userID]; !ok {
		h.users[userID] = make(map[*UserConn]struct{})
	}
	h.users[userID][conn] = struct{}{}
}

func (h *Hub) Unregister(userID uuid.UUID, conn *UserConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.users[userID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.users, userID)
		}
	}
}

func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, e event.Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range userIDs {
		for userConn := range h.users[id] {
			select {
			case userConn.send <- e:
			default:
				close(userConn.done)
			}
		}
	}
}
