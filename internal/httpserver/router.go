package httpserver

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
	router.Use(gin.Logger())
	router.Use(middleware.DummyAuthH())
	//router.Use(middleware.Auth(cfg))

	router.GET("/chats", middleware.CtxTimeout(2*time.Second), chatHandler.GetChats())
	router.POST("/chats", middleware.CtxTimeout(3*time.Second), chatHandler.CreateChat())
	router.GET("/chats/:chatId", middleware.CtxTimeout(2*time.Second), chatHandler.GetChat())
	router.PUT("/chats/:chatId", middleware.CtxTimeout(3*time.Second), chatHandler.UpdateChatName())
	router.DELETE("/chats/:chatId", middleware.CtxTimeout(3*time.Second), chatHandler.DeleteChat())

	router.GET("/chatusers/:chatId", middleware.CtxTimeout(2*time.Second), userHandler.GetChatUsers())
	router.POST("/chatusers/:chatId", middleware.CtxTimeout(3*time.Second), userHandler.AddChatUsers())
	router.PUT("/chatusers/:chatId/:userId", middleware.CtxTimeout(2*time.Second), userHandler.SetChatUserRole())
	router.DELETE("/chatusers/:chatId/:userId", middleware.CtxTimeout(2*time.Second), userHandler.DeleteChatUser())

	router.GET("/chatmessages/:chatId", middleware.CtxTimeout(2*time.Second), messageHandler.GetChatMessages())
	router.GET("/connect/:chatId", middleware.CtxTimeout(2*time.Second), chatHandler.ConnectToChat(wsServer))

	return router
}
