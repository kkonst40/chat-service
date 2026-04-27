package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkonst40/chat-service/internal/app"
	"github.com/kkonst40/chat-service/internal/config"
	"github.com/kkonst40/chat-service/internal/service/logger"
)

func main() {
	slog.SetDefault(slog.Default())

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Config loading", "error", err.Error())
		slog.Info("Server exiting")
		os.Exit(1)
	}

	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("Application creating", "error", err.Error())
		slog.Info("Server exiting")
		os.Exit(1)
	}

	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		err = application.Run(appCtx)
		if err != nil {
			slog.Error("Application running", "error", err.Error())
			slog.Info("Server exiting")
			os.Exit(1)
		}
	}()

	<-appCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(shutdownCtx)
}
