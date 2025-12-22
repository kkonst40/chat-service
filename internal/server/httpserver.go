package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/repository"
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

	chatRepository := repository.NewInMemoryChatRepository()
	messageRepository := repository.NewInMemoryMessageRepository()
	chatService := service.NewChatService(chatRepository)
	messageService := service.NewMessageService(messageRepository, chatService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	ws := ws.NewWsServer(chatService, messageService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	//router.Use(middleware.AuthMiddleware(jwtConfig))
	//router.Use(middleware.DummyMiddleware())
	//router.GET("/1", func(c *gin.Context) {
	//	c.File("./static/chatlist1.html")
	//})
	//router.GET("/2", func(c *gin.Context) {
	//	c.File("./static/chatlist2.html")
	//})

	router.GET("/chats", chatHandler.GetChats())
	router.POST("/chats", chatHandler.CreateChat())
	router.GET("/chats/:id", chatHandler.GetChat())
	router.PUT("/chats/:id", chatHandler.UpdateChatName())
	router.DELETE("/chats/:id", chatHandler.DeleteChat())
	router.POST("/connect/:id", chatHandler.ConnectToChat(ws))

	router.GET("/chatmessages/:id", messageHandler.GetChatMessages())

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
