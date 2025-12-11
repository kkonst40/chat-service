package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/handlers"
	"github.com/kkonst40/ichat/internal/repositories"
	"github.com/kkonst40/ichat/internal/services"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	chatHandler, messageHandler := initHandlers()

	router := gin.Default()
	router.GET("/chats", chatHandler.GetChats())
	router.POST("/chats", chatHandler.CreateChat())
	router.GET("/chats/:id", chatHandler.GetChat())

	router.GET("/chatmessages/:id", messageHandler.GetChatMessages())
	router.POST("/messages", messageHandler.SendMessages())

	router.Run("localhost:8080")
}

func initHandlers() (*handlers.ChatHandler, *handlers.MessageHandler) {
	var chatRepository repositories.ChatRepository = repositories.NewInMemoryChatRepository()
	var messageRepository repositories.MessageRepository = repositories.NewInMemoryMessageRepository()

	chatService := services.NewChatService(chatRepository)
	messageService := services.NewMessageService(messageRepository, chatService)

	chatHandler := handlers.NewChatHandler(chatService)
	messageHandler := handlers.NewMessageHandler(messageService)

	return chatHandler, messageHandler
}
