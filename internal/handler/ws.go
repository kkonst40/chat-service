package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kkonst40/ichat/internal/hub"
)

type WSHandler struct {
	hub *hub.Hub
}

func NewWSHandler(hub *hub.Hub) *WSHandler {
	return &WSHandler{
		hub: hub,
	}
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

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	userConn := hub.NewUserConn(conn)

	h.hub.ServeConn(requesterID, userConn)
}
