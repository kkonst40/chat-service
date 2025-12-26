package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/middleware"
	"github.com/kkonst40/ichat/internal/repository/memory"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"
)

type HttpServer struct {
	router         *gin.Engine
	ws             *ws.Server
	address        string
	chatHandler    *handler.ChatHandler
	messageHandler *handler.MessageHandler
}

func NewHttpServer() *HttpServer {
	gin.SetMode(gin.ReleaseMode)

	_, err := config.LoadJwtConfig()
	if err != nil {
		panic(err)
	}

	userRepository := memory.NewUserRepository()
	chatRepository := memory.NewChatRepository()
	messageRepository := memory.NewMessageRepository()

	userService := service.NewUserService(userRepository)
	chatService := service.NewChatService(chatRepository, userService)
	messageService := service.NewMessageService(messageRepository, chatService, userService)

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	ws := ws.NewWsServer(chatService, messageService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	//router.Use(middleware.AuthMiddleware(jwtConfig))
	router.Use(middleware.DummyMiddleware())
	//router.GET("/1", func(c *gin.Context) {
	//	c.File("./static/user1.html")
	//})
	//router.GET("/2", func(c *gin.Context) {
	//	c.File("./static/user2.html")
	//})

	router.GET("/chats", chatHandler.GetChats())
	router.POST("/chats", chatHandler.CreateChat())
	router.GET("/chats/:chaId", chatHandler.GetChat())
	router.PUT("/chats/:chaId", chatHandler.UpdateChatName())
	router.DELETE("/chats/:chaId", chatHandler.DeleteChat())
	router.GET("/connect/:chaId", chatHandler.ConnectToChat(ws))

	router.GET("/chatusers/:chatId", userHandler.GetChatUsers())
	router.POST("/chatusers/:chatId", userHandler.AddChatUsers())
	router.PUT("/chatusers/:chatId/:userId", userHandler.SetChatUserRole())
	router.DELETE("/chatusers/:chatId/:userId", userHandler.DeleteChatUser())

	router.GET("/chatmessages/:chatId", messageHandler.GetChatMessages())

	server := &HttpServer{
		router:         router,
		ws:             ws,
		address:        "localhost:8080",
		chatHandler:    chatHandler,
		messageHandler: messageHandler,
	}

	return server
}

func (s *HttpServer) Run() {
	s.router.Run(s.address)
}
