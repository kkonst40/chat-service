package dispatcher

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/event"
	"github.com/kkonst40/chat-service/internal/hub"
)

type Dispatcher struct {
	WsHub    *hub.Hub
	userRepo UserRepository
}

type UserRepository interface {
	GetChatUserIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error)
}

func New(
	wsHub *hub.Hub,
	userRepo UserRepository,
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
