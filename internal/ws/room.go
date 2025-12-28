package ws

import (
	"context"
	"sync"
)

type room struct {
	ctx    context.Context
	cancel context.CancelFunc

	users      map[*user]bool
	addUser    chan *user
	removeUser chan *user
	broadcast  chan message
	mutex      sync.RWMutex
}

func newRoom(ctxParent context.Context) *room {
	ctx, cancel := context.WithCancel(ctxParent)

	return &room{
		ctx:        ctx,
		cancel:     cancel,
		users:      make(map[*user]bool),
		addUser:    make(chan *user),
		removeUser: make(chan *user),
		broadcast:  make(chan message),
	}
}
