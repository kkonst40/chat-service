package dispatcher

import (
	"context"

	"github.com/kkonst40/ichat/internal/domain/event"
	"github.com/kkonst40/ichat/internal/hub"
	"github.com/kkonst40/ichat/internal/repository"
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

func (d *Dispatcher) Publish(e event.Event) {
	userIDs, _ := d.userRepo.GetChatUserIDs(context.TODO(), e.ChatID)

	d.WsHub.BroadcastToUsers(userIDs, e)
}
