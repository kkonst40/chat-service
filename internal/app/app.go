package app

import (
	"database/sql"
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

func New(cfg *config.JWTConfig) (*App, error) {
	// host: localhost
	// user: app_user
	// password: app_password
	// dbname: app_db
	dsn := "postgres://app_user:app_password@localhost:5432/app_db?sslmode=disable"
	//dsn := "postgres://app_user:app_password@postgres:5432/app_db"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Ошибка при создании объекта db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)                 // Максимум открытых соединений
	db.SetMaxIdleConns(5)                  // Сколько держать в простое
	db.SetConnMaxLifetime(5 * time.Minute) // Пересоздавать соединения каждые 5 минут

	if err := db.Ping(); err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Успешное подключение к базе данных")

	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	log.Printf("Репозитории инициализированы: %v, %v, %v\n", userRepo, chatRepo, messageRepo)

	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userService)
	messageService := service.NewMessageService(messageRepo, chatService, userService)

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	wsServer := ws.NewWsServer(chatService, messageService)

	router := httpserver.NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsServer,
		cfg,
	)

	server := httpserver.New(router, "localhost:8080")

	return &App{
		httpServer: server,
	}, nil
}

func (a *App) Run() error {
	return a.httpServer.Run()
}
