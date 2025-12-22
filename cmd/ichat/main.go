package main

import (
	"github.com/kkonst40/ichat/internal/server"
)

func main() {
	httpServer := server.NewHttpServer()
	httpServer.Run()
}
