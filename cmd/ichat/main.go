package main

import (
	"log/slog"
	"os"

	"github.com/kkonst40/ichat/internal/app"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/kkonst40/ichat/internal/logger"
)

func main() {
	log := logger.New("dev")
	slog.SetDefault(log)

	cfg, err := config.Load("dev")
	if err != nil {
		log.Error("config loading error", "error", err.Error())
		log.Info("Server exiting")
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Error("application creating error", "error", err.Error())
		log.Info("Server exiting")
		os.Exit(1)
	}

	err = application.Run()
	if err != nil {
		log.Error("application running error", "error", err.Error())
		log.Info("Server exiting")
		os.Exit(1)
	}
}
