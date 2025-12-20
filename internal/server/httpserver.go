package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/middleware"
	"github.com/kkonst40/ichat/internal/repository"
	"github.com/kkonst40/ichat/internal/service"
)

type HttpServer struct {
	router         *gin.Engine
	address        string
	chatHandler    *handler.ChatHandler
	messageHandler *handler.MessageHandler
}

func NewHttpServer() *HttpServer {
	gin.SetMode(gin.ReleaseMode)

	jwtConfig, err := config.LoadJwtConfig()
	if err != nil {
		panic(err)
	}

	chatRepository := repository.NewInMemoryChatRepository()
	messageRepository := repository.NewInMemoryMessageRepository()
	chatService := service.NewChatService(chatRepository)
	messageService := service.NewMessageService(messageRepository, chatService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.AuthMiddleware(jwtConfig))

	router.GET("/chats", middleware.AuthMiddleware(jwtConfig), chatHandler.GetChats())
	router.POST("/chats", middleware.AuthMiddleware(jwtConfig), chatHandler.CreateChat())
	router.GET("/chats/:id", middleware.AuthMiddleware(jwtConfig), chatHandler.GetChat())
	router.PUT("/chats/:id", middleware.AuthMiddleware(jwtConfig), chatHandler.UpdateChatName())
	router.DELETE("/chats/:id", middleware.AuthMiddleware(jwtConfig), chatHandler.DeleteChat())

	router.GET("/chatmessages/:id", middleware.AuthMiddleware(jwtConfig), messageHandler.GetChatMessages())
	router.POST("/messages", middleware.AuthMiddleware(jwtConfig), messageHandler.SendMessages())

	server := &HttpServer{
		router:         router,
		address:        "localhost:8080",
		chatHandler:    chatHandler,
		messageHandler: messageHandler,
	}

	return server
}

func (s *HttpServer) Run() {
	s.router.Run(s.address)
}
