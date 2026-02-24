package ws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/service"
)

type room struct {
	ctx            context.Context
	cancel         context.CancelFunc
	messageService *service.MessageService

	users      map[*user]bool
	addUser    chan *user
	removeUser chan *user
	eventQueue chan roomEvent
}

func newRoom(ctxParent context.Context, messageService *service.MessageService) *room {
	ctx, cancel := context.WithCancel(ctxParent)

	return &room{
		ctx:            ctx,
		cancel:         cancel,
		messageService: messageService,

		users:      make(map[*user]bool),
		addUser:    make(chan *user),
		removeUser: make(chan *user),
		eventQueue: make(chan roomEvent),
	}
}

func (r *room) run(chatID uuid.UUID) {
	for {
		select {
		case <-r.ctx.Done():
			slog.Debug("Room context done", "roomID", chatID)
			return

		case user := <-r.addUser:
			r.users[user] = true

		case user := <-r.removeUser:
			if _, ok := r.users[user]; ok {
				delete(r.users, user)
				close(user.send)
			}
			if len(r.users) == 0 {
				return
			}

		case event := <-r.eventQueue:
			event, err := r.handleEvent(event, chatID)
			if err != nil {
				slog.Error("Handling event error", "errors", err.Error())
				continue
			}

			for user := range r.users {
				select {
				case user.send <- event:
				default:
					delete(r.users, user)
					close(user.send)
				}
			}
		}
	}
}

func (r *room) handleEvent(event roomEvent, chatID uuid.UUID) (roomEvent, error) {
	switch event.Type {
	case ActionCreate:
		msgID, err := uuid.NewV7()
		if err != nil {
			return roomEvent{}, fmt.Errorf("generating msg id error")
		}

		event.MsgID = msgID

		go func() {
			_, err = r.messageService.CreateMessage(
				r.ctx,
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
		err := r.messageService.UpdateMessage(
			r.ctx,
			event.MsgID,
			event.Text,
			event.UserID,
		)

		if err != nil {
			return roomEvent{}, err
		}

	case ActionDelete:
		err := r.messageService.DeleteMessage(
			r.ctx,
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
