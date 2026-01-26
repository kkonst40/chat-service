package main

import (
	"log/slog"
	"os"

	"github.com/kkonst40/ichat/internal/app"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/logger"
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

	err = application.Run()
	if err != nil {
		slog.Error("Application running", "error", err.Error())
		slog.Info("Server exiting")
		os.Exit(1)
	}
}
