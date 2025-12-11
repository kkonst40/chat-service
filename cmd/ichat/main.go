package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/handlers"
	"github.com/kkonst40/ichat/internal/repositories"
	"github.com/kkonst40/ichat/internal/services"
)

func main() {
	server := Server{}
	server.Build()
	server.Run()
}

type Server struct {
	router         *gin.Engine
	address        string
	chatHandler    *handlers.ChatHandler
	messageHandler *handlers.MessageHandler
}

func (s *Server) Build() {
	gin.SetMode(gin.ReleaseMode)
	s.initHandlers()
	s.address = "localhost:8080"

	router := gin.Default()
	router.GET("/chats", s.chatHandler.GetChats())
	router.POST("/chats", s.chatHandler.CreateChat())
	router.GET("/chats/:id", s.chatHandler.GetChat())

	router.GET("/chatmessages/:id", s.messageHandler.GetChatMessages())
	router.POST("/messages", s.messageHandler.SendMessages())
}

func (s *Server) Run() {
	s.router.Run("localhost:8080")
}

func (s *Server) initHandlers() {
	chatRepository := repositories.NewInMemoryChatRepository()
	messageRepository := repositories.NewInMemoryMessageRepository()

	chatService := services.NewChatService(chatRepository)
	messageService := services.NewMessageService(messageRepository, chatService)

	s.chatHandler = handlers.NewChatHandler(chatService)
	s.messageHandler = handlers.NewMessageHandler(messageService)
}
