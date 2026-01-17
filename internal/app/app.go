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
	httpServer *http.Server
	wsServer   *ws.Server
	db         *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := NewDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	slog.Info("Successful connection to the database")

	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	// for test
	//memDB := memory.NewDB()
	//userRepo := memory.NewUserRepository(memDB)
	//chatRepo := memory.NewChatRepository(memDB)
	//messageRepo := memory.NewMessageRepository(memDB)

	slog.Info("Repositories are initialized")

	userService := service.NewUserService(userRepo, cfg.SSOURL)
	chatService := service.NewChatService(chatRepo, userService)
	messageService := service.NewMessageService(messageRepo, chatService, userService)

	slog.Info("Services are initialized")

	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)
	messageHandler := handler.NewMessageHandler(messageService)

	slog.Info("Handlers are initialized")

	wsServer := ws.NewServer(chatService, messageService)

	slog.Info("WebSocket server is initialized")

	router := NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsServer,
		cfg,
	)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return &App{
		httpServer: httpServer,
		wsServer:   wsServer,
		db:         db,
	}, nil
}

func (a *App) Run() error {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()
	slog.Info("Server started", "address", a.httpServer.Addr)

	<-appCtx.Done()
	slog.Warn("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.wsServer.Shutdown()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.Warn("Server forced to shutdown", "error", err.Error())
	}
	if err := a.db.Close(); err != nil {
		slog.Warn("DB close error", "error", err.Error())
	}

	slog.Warn("Server exiting")
	return nil
}

func NewDB(user, pwd, host, dbName string) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%v:%v@%v/%v",
		user, pwd, host, dbName)

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

	return db, nil
}
