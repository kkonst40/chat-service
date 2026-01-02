package app

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/httpserver"
	"github.com/kkonst40/ichat/internal/repository/postgres"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	httpServer *httpserver.Server
}

func New(cfg *config.Config) (*App, error) {
	dbUrl := fmt.Sprintf(
		"postgres://%v:%v@%v/%v",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.DBName,
	)
	//if strings.HasPrefix(cfg.DB.Host, "localhost") {
	//	dbUrl += "?sslmode=disabled"
	//}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("Error creating db object: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to the database: %v", err)
	}

	log.Println("Successful connection to the database")

	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	log.Println("Repositories are initialized")

	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userService)
	messageService := service.NewMessageService(messageRepo, chatService, userService)

	log.Println("Services are initialized")

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	log.Println("Handlers are initialized")

	wsServer := ws.NewWsServer(chatService, messageService)

	router := httpserver.NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsServer,
		cfg,
	)

	server := httpserver.New(router, "localhost:8080")

	log.Println("Server is initialized")

	return &App{
		httpServer: server,
	}, nil
}

func (a *App) Run() error {
	return a.httpServer.Run()
}
