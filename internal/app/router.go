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

	// for test
	router.GET("/1", func(c *gin.Context) {
		c.File("static/cd1.html")
	})
	router.GET("/2", func(c *gin.Context) {
		c.File("static/cd2.html")
	})

	router.GET("/connect/:chatId", middleware.DummyAuthQ(), chatHandler.ConnectToChat(wsServer))

	http := router.Group("/")
	http.Use(middleware.CtxTimeout(3 * time.Second))
	{
		http.Use(middleware.DummyAuthH())

		http.GET("/chats", chatHandler.GetChats())
		http.POST("/chats", chatHandler.CreateChat())
		http.GET("/chats/:chatId", chatHandler.GetChat())
		http.PUT("/chats/:chatId", chatHandler.UpdateChatName())
		http.DELETE("/chats/:chatId", chatHandler.DeleteChat())

		http.GET("/chatusers/:chatId", userHandler.GetChatUsers())
		http.POST("/chatusers/:chatId", userHandler.AddChatUsers())
		http.PUT("/chatusers/:chatId/:userId", userHandler.UpdateChatUserRole())
		http.DELETE("/chatusers/:chatId/:userId", userHandler.DeleteChatUser())

		http.GET("/chatmessages/:chatId", messageHandler.GetChatMessages())
	}

	return router
}
