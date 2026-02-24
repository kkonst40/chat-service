package app

import (
	"net/http"
	"time"

	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/middleware"
	"github.com/kkonst40/ichat/internal/ws"
)

func NewRouter(
	chatHandler *handler.ChatHandler,
	userHandler *handler.UserHandler,
	messageHandler *handler.MessageHandler,
	wsServer *ws.Server,
	cfg *config.Config,
) http.Handler {
	httpRouter := http.NewServeMux()

	httpRouter.HandleFunc("GET /chats", chatHandler.GetChats)
	httpRouter.HandleFunc("POST /chats", chatHandler.CreateChat)
	httpRouter.HandleFunc("GET /chats/{chatId}", chatHandler.GetChat)
	httpRouter.HandleFunc("PUT /chats/{chatId}", chatHandler.UpdateChatName)
	httpRouter.HandleFunc("DELETE /chats/{chatId}", chatHandler.DeleteChat)

	httpRouter.HandleFunc("GET /chatusers/{chatId}", userHandler.GetChatUsers)
	httpRouter.HandleFunc("POST /chatusers/{chatId}", userHandler.AddChatUsers)
	httpRouter.HandleFunc("PUT /chatusers/{chatId}/{userId}", userHandler.UpdateChatUserRole)
	httpRouter.HandleFunc("DELETE /chatusers/{chatId}/{userId}", userHandler.DeleteChatUser)

	httpRouter.HandleFunc("GET /chatmessages/{chatId}", messageHandler.GetChatMessages)

	// for test
	httpRouter.HandleFunc("GET /chatlist", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/chatlist.html")
	})
	httpRouter.HandleFunc("GET /chatroom", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/chatroom.html")
	})
	//

	wsRouter := http.NewServeMux()
	wsRouter.HandleFunc("GET /connect/{chatId}", chatHandler.ConnectToChat(wsServer))

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
	mainRouter.Handle("/", httpStack(httpRouter))
	mainRouter.Handle("/", wsStack(wsRouter))

	return mainRouter
}
