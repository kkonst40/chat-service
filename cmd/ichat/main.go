package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/handlers"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	chatHandler := handlers.NewChatHandler()
	router := gin.Default()
	router.GET("/chats", chatHandler.GetChats())
	router.GET("/chats/:id", chatHandler.GetChat())
	router.POST("/chats", chatHandler.CreateChat())

	router.Run("localhost:8080")
}
