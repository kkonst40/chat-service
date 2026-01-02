package main

import (
	"log"

	"github.com/kkonst40/ichat/internal/app"
	"github.com/kkonst40/ichat/internal/config"
)

func main() {
	cfg, err := config.Load("dev")
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
