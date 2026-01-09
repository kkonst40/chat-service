package ws

import (
	"context"
)

type room struct {
	ctx    context.Context
	cancel context.CancelFunc

	users      map[*user]bool
	addUser    chan *user
	removeUser chan *user
	eventQueue chan roomEvent
	broadcast  chan roomEvent
}

func newRoom(ctxParent context.Context) *room {
	ctx, cancel := context.WithCancel(ctxParent)

	return &room{
		ctx:        ctx,
		cancel:     cancel,
		users:      make(map[*user]bool),
		addUser:    make(chan *user),
		removeUser: make(chan *user),
		eventQueue: make(chan roomEvent),
		broadcast:  make(chan roomEvent),
	}
}
