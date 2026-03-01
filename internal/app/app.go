package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/dispatcher"
	pb "github.com/kkonst40/ichat/internal/gen/user"
	"github.com/kkonst40/ichat/internal/handler"
	"github.com/kkonst40/ichat/internal/integration/sso"
	"github.com/kkonst40/ichat/internal/repository/postgres"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	httpServer *http.Server
	wsServer   *ws.Server
	db         *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := SetupDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	slog.Info("Successful connection to the database")

	var (
		userRepo    = postgres.NewUserRepository(db)
		chatRepo    = postgres.NewChatRepository(db)
		messageRepo = postgres.NewMessageRepository(db)
	)

	// for test
	// var (
	// 	memDB       = memory.NewDB()
	// 	userRepo    = memory.NewUserRepository(memDB)
	// 	chatRepo    = memory.NewChatRepository(memDB)
	// 	messageRepo = memory.NewMessageRepository(memDB)
	// )

	slog.Info("Repositories are initialized")

	conn, err := grpc.NewClient(
		cfg.SSOAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	var (
		ssoClient  = sso.NewSSOClient(pb.NewUserServiceClient(conn))
		dispatcher = dispatcher.New(nil, userRepo)

		userService    = service.NewUserService(userRepo, dispatcher, ssoClient)
		chatService    = service.NewChatService(chatRepo, userService, dispatcher)
		messageService = service.NewMessageService(messageRepo, chatService, userService, dispatcher, 4096)
	)
	slog.Info("Services are initialized")

	var (
		validator      = handler.NewValidator()
		userHandler    = handler.NewUserHandler(userService, validator)
		chatHandler    = handler.NewChatHandler(chatService, validator)
		messageHandler = handler.NewMessageHandler(messageService, validator)
	)
	slog.Info("Handlers are initialized")

	wsServer := ws.NewServer()
	//temporary
	dispatcher.WsServer = wsServer

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

	slog.Info("HTTP server is initialized")

	return &App{
		httpServer: httpServer,
		wsServer:   wsServer,
		db:         db,
	}, nil
}

func (a *App) Run() error {
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP serve error: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err.Error())
	}

	if err := a.db.Close(); err != nil {
		slog.Error("DB close error", "error", err.Error())
	}
}

func SetupDB(user, pwd, host, dbName string) (*sql.DB, error) {
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
