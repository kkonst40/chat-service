package app

import (
	"time"

	"github.com/gin-gonic/gin"
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
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Error())

	router.GET("/connect/:chatId", middleware.DummyAuthH(), chatHandler.ConnectToChat(wsServer))

	fast := router.Group("/")
	fast.Use(middleware.CtxTimeout(2 * time.Second))
	{
		fast.Use(middleware.DummyAuthH())

		fast.GET("/chats", chatHandler.GetChats())
		fast.GET("/chats/:chatId", chatHandler.GetChat())

		fast.GET("/chatusers/:chatId", userHandler.GetChatUsers())
		fast.PUT("/chatusers/:chatId/:userId", userHandler.UpdateChatUserRole())
		fast.DELETE("/chatusers/:chatId/:userId", userHandler.DeleteChatUser())

		fast.GET("/chatmessages/:chatId", messageHandler.GetChatMessages())
	}

	slow := router.Group("/")
	slow.Use(middleware.CtxTimeout(3 * time.Second))
	{
		slow.Use(middleware.DummyAuthH())

		slow.POST("/chats", chatHandler.CreateChat())
		slow.PUT("/chats/:chatId", chatHandler.UpdateChatName())
		slow.DELETE("/chats/:chatId", chatHandler.DeleteChat())

		slow.POST("/chatusers/:chatId", userHandler.AddChatUsers())
	}

	return router
}
