package dispatcher

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/event"
	"github.com/kkonst40/chat-service/internal/hub"
	"github.com/kkonst40/chat-service/internal/repository"
)

type Dispatcher struct {
	WsHub    *hub.Hub
	userRepo repository.UserRepository
}

func New(
	wsHub *hub.Hub,
	userRepo repository.UserRepository,
) *Dispatcher {
	return &Dispatcher{
		WsHub:    wsHub,
		userRepo: userRepo,
	}
}

func (d *Dispatcher) Publish(e event.Event, userIDs ...uuid.UUID) {
	if len(userIDs) == 0 {
		userIDs, _ = d.userRepo.GetChatUserIDs(context.TODO(), e.ChatID)
	}

	d.WsHub.BroadcastToUsers(userIDs, e)
}
