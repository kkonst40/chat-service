package ws

import (
	"sync"
)

type room struct {
	users      map[*user]bool
	addUser    chan *user
	removeUser chan *user
	broadcast  chan message
	mutex      sync.RWMutex
}

func newRoom() *room {
	return &room{
		users:      make(map[*user]bool),
		addUser:    make(chan *user),
		removeUser: make(chan *user),
		broadcast:  make(chan message),
	}
}
