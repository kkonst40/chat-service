package dispatcher

import (
	"context"

	"github.com/kkonst40/ichat/internal/domain/event"
	"github.com/kkonst40/ichat/internal/repository"
	"github.com/kkonst40/ichat/internal/ws"
)

type Dispatcher struct {
	WsServer *ws.Server
	userRepo repository.UserRepository
}

func New(
	wsServer *ws.Server,
	userRepo repository.UserRepository,
) *Dispatcher {
	return &Dispatcher{
		WsServer: wsServer,
		userRepo: userRepo,
	}
}

func (d *Dispatcher) Publish(e event.Event) {
	userIDs, _ := d.userRepo.GetChatUserIDs(context.TODO(), e.ChatID)

	d.WsServer.SendEvent(userIDs, e)
}
