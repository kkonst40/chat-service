package app

import (
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/httpserver"
	"github.com/kkonst40/ichat/internal/repository/memory"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"
)

type App struct {
	httpServer *httpserver.Server
}

func New(cfg *config.JWTConfig) (*App, error) {
	userRepo := memory.NewUserRepository()
	chatRepo := memory.NewChatRepository()
	messageRepo := memory.NewMessageRepository()

	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userService)
	messageService := service.NewMessageService(messageRepo, chatService, userService)

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	wsServer := ws.NewWsServer(chatService, messageService)

	router := httpserver.NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsServer,
		cfg,
	)

	server := httpserver.New(router, "localhost:8080")

	return &App{
		httpServer: server,
	}, nil
}

func (a *App) Run() error {
	return a.httpServer.Run()
}
