package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handlers"
	"github.com/kkonst40/ichat/internal/middleware"
	"github.com/kkonst40/ichat/internal/repositories"
	"github.com/kkonst40/ichat/internal/services"
)

type Server struct {
	router         *gin.Engine
	address        string
	chatHandler    *handlers.ChatHandler
	messageHandler *handlers.MessageHandler
}

func New() *Server {
	gin.SetMode(gin.ReleaseMode)

	jwtConfig, err := config.LoadJwtConfig()
	if err != nil {
		panic(err)
	}

	chatRepository := repositories.NewInMemoryChatRepository()
	messageRepository := repositories.NewInMemoryMessageRepository()
	chatService := services.NewChatService(chatRepository)
	messageService := services.NewMessageService(messageRepository, chatService)
	chatHandler := handlers.NewChatHandler(chatService)
	messageHandler := handlers.NewMessageHandler(messageService)

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

	server := &Server{
		router:         router,
		address:        "localhost:8080",
		chatHandler:    chatHandler,
		messageHandler: messageHandler,
	}

	return server
}

func (s *Server) Run() {
	s.router.Run(s.address)
}
