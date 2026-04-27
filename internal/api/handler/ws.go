package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kkonst40/chat-service/internal/auth"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
	"github.com/kkonst40/chat-service/internal/hub"
	"github.com/kkonst40/chat-service/internal/limit/conntracker"
)

type WSHandler struct {
	hub         *hub.Hub
	connTracker *conntracker.ConnTracker
}

func NewWSHandler(hub *hub.Hub, connTracker *conntracker.ConnTracker) *WSHandler {
	return &WSHandler{
		hub:         hub,
		connTracker: connTracker,
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
	requesterID := auth.GetUserID(ctx)

	clientIP := GetRealIP(r)

	if !h.connTracker.Acquire(clientIP) {
		WriteError(ctx, w, fmt.Errorf("%w: from IP %s", errs.ErrTooManyOpenConnections, clientIP))
		return
	}
	defer h.connTracker.Release(clientIP)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return
	}

	userConn := hub.NewUserConn(conn)

	h.hub.ServeConn(requesterID, userConn)
}
