package app

import (
	"net/http"
	"time"

	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/middleware"
)

func NewRouter(
	chatHandler *handler.ChatHandler,
	userHandler *handler.UserHandler,
	messageHandler *handler.MessageHandler,
	wsHandler *handler.WSHandler,
	cfg *config.Config,
) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /chats", chatHandler.GetChats)
	router.HandleFunc("POST /chats/personal", chatHandler.CreatePersonalChat)
	router.HandleFunc("POST /chats/group", chatHandler.CreateGroupChat)
	router.HandleFunc("GET /chats/{chatId}", chatHandler.GetChat)
	router.HandleFunc("PUT /chats/{chatId}", chatHandler.UpdateChatName)
	router.HandleFunc("DELETE /chats/{chatId}", chatHandler.DeleteChat)

	router.HandleFunc("GET /chatusers/{chatId}", userHandler.GetChatUsers)
	router.HandleFunc("POST /chatusers/{chatId}", userHandler.AddChatUsers)
	router.HandleFunc("PUT /chatusers/{chatId}/{userId}", userHandler.UpdateChatUserRole)
	router.HandleFunc("DELETE /chatusers/{chatId}/{userId}", userHandler.DeleteChatUser)

	router.HandleFunc("GET /chatmessages/{chatId}", messageHandler.GetChatMessages)
	router.HandleFunc("POST /chatmessages/{chatId}", messageHandler.CreateMessage)
	router.HandleFunc("PUT /chatmessages/{msgId}", messageHandler.UpdateMessage)
	router.HandleFunc("DELETE /chatmessages/{msgId}", messageHandler.DeleteMessage)

	// for test
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	httpStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Logger,
		middleware.Timeout(3*time.Second),
		middleware.Auth(cfg),
	)

	wsStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Logger,
		middleware.Auth(cfg),
	)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/", httpStack(router))
	mainRouter.Handle("GET /connect", wsStack(http.HandlerFunc(wsHandler.HandleConnection)))

	return mainRouter
}
