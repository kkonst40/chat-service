package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/repository/postgres"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	server *http.Server
	db     *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	dbUrl := fmt.Sprintf("postgres://%v:%v@%v/%v",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DBName)

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

	slog.Info("Successful connection to the database")

	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	slog.Info("Repositories are initialized")

	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userService)
	messageService := service.NewMessageService(messageRepo, chatService, userService)

	slog.Info("Services are initialized")

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	slog.Info("Handlers are initialized")

	wsServer := ws.NewWsServer(chatService, messageService)

	router := NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsServer,
		cfg,
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return &App{
		server: server,
		db:     db,
	}, nil
}

func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()
	slog.Info("Server started on :8080")

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Server forced to shutdown: %v", err)
	}
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("DB close error: %v", err)
	}

	slog.Info("Server exiting")
	return nil
}
