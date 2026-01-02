package httpserver

import (
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
	cfg *config.JWTConfig,
) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.Dummy())
	router.Use(middleware.Auth(cfg))

	router.GET("/chats", chatHandler.GetChats())
	router.POST("/chats", chatHandler.CreateChat())
	router.GET("/chats/:chatId", chatHandler.GetChat())
	router.PUT("/chats/:chatId", chatHandler.UpdateChatName())
	router.DELETE("/chats/:chatId", chatHandler.DeleteChat())
	router.GET("/connect/:chatId", chatHandler.ConnectToChat(wsServer))

	router.GET("/chatusers/:chatId", userHandler.GetChatUsers())
	router.POST("/chatusers/:chatId", userHandler.AddChatUsers())
	router.PUT("/chatusers/:chatId/:userId", userHandler.SetChatUserRole())
	router.DELETE("/chatusers/:chatId/:userId", userHandler.DeleteChatUser())

	router.GET("/chatmessages/:chatId", messageHandler.GetChatMessages())

	return router
}
