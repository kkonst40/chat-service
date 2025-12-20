package main

import (
	"github.com/kkonst40/ichat/internal/server"
)

func main() {
	s := server.NewHttpServer()
	s.Run()
}
