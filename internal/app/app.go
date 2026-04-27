package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kkonst40/chat-service/internal/api"
	"github.com/kkonst40/chat-service/internal/api/handler"
	"github.com/kkonst40/chat-service/internal/api/limit/conntracker"
	"github.com/kkonst40/chat-service/internal/api/limit/ratelimiter"
	"github.com/kkonst40/chat-service/internal/config"
	pb "github.com/kkonst40/chat-service/internal/gen/user"
	"github.com/kkonst40/chat-service/internal/hub"
	"github.com/kkonst40/chat-service/internal/repository/postgres"
	"github.com/kkonst40/chat-service/internal/service"
	"github.com/kkonst40/chat-service/internal/service/auth"
	"github.com/kkonst40/chat-service/internal/service/cache"
	"github.com/kkonst40/chat-service/internal/service/dispatcher"
	"github.com/kkonst40/chat-service/internal/service/eventbus"
	"github.com/kkonst40/chat-service/internal/service/integration/sso"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	server        *http.Server
	eventConsumer *eventbus.Consumer
	db            *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := SetupDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}
	slog.Info("Successful connection to the database")

	redisClient, err := SetupRedis(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return nil, err
	}
	slog.Info("Successful connection to the Redis")

	var (
		userRepo    = postgres.NewUserRepository(db)
		chatRepo    = postgres.NewChatRepository(db)
		messageRepo = postgres.NewMessageRepository(db)
	)

	slog.Info("Repositories are initialized")

	conn, err := grpc.NewClient(
		cfg.SSOAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	var (
		userLoginCache = cache.NewRedisUserLoginCache(redisClient, time.Duration(cfg.LoginCacheTTLHours)*time.Hour)
		ssoClient      = sso.NewClient(pb.NewUserServiceClient(conn))
		wsHub          = hub.NewHub()
		dispatcher     = dispatcher.New(wsHub, userRepo)
		tokenValidator = auth.NewTokenValidator(cfg)
		rateLimiter    = ratelimiter.New(cfg)
		connTracker    = conntracker.New(cfg.WSConnsPerIP)

		eventConsumer = eventbus.NewConsumer(cfg, userLoginCache)
	)

	var (
		userService    = service.NewUserService(userRepo, dispatcher, ssoClient, userLoginCache)
		chatService    = service.NewChatService(chatRepo, userService, dispatcher)
		messageService = service.NewMessageService(messageRepo, chatService, userService, dispatcher, 4096)
	)
	slog.Info("Services are initialized")

	var (
		validator      = handler.NewValidator()
		userHandler    = handler.NewUserHandler(userService, validator)
		chatHandler    = handler.NewChatHandler(chatService, validator)
		messageHandler = handler.NewMessageHandler(messageService, validator)
		wsHandler      = handler.NewWSHandler(wsHub, connTracker)
	)
	slog.Info("Handlers are initialized")

	router := api.NewRouter(
		chatHandler,
		userHandler,
		messageHandler,
		wsHandler,
		tokenValidator,
		rateLimiter,
		cfg,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	slog.Info("HTTP server is initialized")

	return &App{
		server:        server,
		eventConsumer: eventConsumer,
		db:            db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errChan := make(chan error, 2)

	go func() {
		if err := a.eventConsumer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("Event bus consumer error: %w", err)
		}
	}()

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP serve error: %w", err)
		}
	}()

	return <-errChan
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err.Error())
	}

	if err := a.eventConsumer.Close(); err != nil {
		slog.Error("Event consumer close error", "error", err.Error())
	}

	if err := a.db.Close(); err != nil {
		slog.Error("DB close error", "error", err.Error())
	}
}

func SetupDB(user, pwd, host, port, dbName string) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, pwd, host, port, dbName)

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error creating db object: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	return db, nil
}

func SetupRedis(host, port, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to the Redis: %w", err)
	}

	return client, nil
}
