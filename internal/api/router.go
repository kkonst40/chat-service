package api

import (
	"net/http"
	"time"

	"github.com/kkonst40/chat-service/internal/api/handler"
	"github.com/kkonst40/chat-service/internal/api/middleware"
	"github.com/kkonst40/chat-service/internal/auth"
	"github.com/kkonst40/chat-service/internal/config"
	"github.com/kkonst40/chat-service/internal/limit/ratelimiter"
)

func NewRouter(
	chatHandler *handler.ChatHandler,
	userHandler *handler.UserHandler,
	messageHandler *handler.MessageHandler,
	wsHandler *handler.WSHandler,
	tokenValidator *auth.TokenValidator,
	rateLimiter *ratelimiter.IPRateLimiter,
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

	router.HandleFunc("GET /chats/{chatId}/messages", messageHandler.GetChatMessages)
	router.HandleFunc("POST /chats/{chatId}/messages", messageHandler.CreateMessage)
	router.HandleFunc("PUT /messages/{msgId}", messageHandler.UpdateMessage)
	router.HandleFunc("DELETE /messages/{msgId}", messageHandler.DeleteMessage)

	httpStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Logger,
		middleware.LimitRate(rateLimiter),
		middleware.Timeout(time.Duration(cfg.RequestTimeoutSeconds)*time.Second),
		middleware.Auth(tokenValidator, cfg.JWT.CookieName),
	)

	wsStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Logger,
		middleware.LimitRate(rateLimiter),
		middleware.Auth(tokenValidator, cfg.JWT.CookieName),
	)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/", httpStack(router))
	mainRouter.Handle("GET /connect", wsStack(http.HandlerFunc(wsHandler.HandleConnection)))

	return mainRouter
}
