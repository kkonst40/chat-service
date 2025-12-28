package httpserver

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	router  *gin.Engine
	address string
}

func New(router *gin.Engine, address string) *Server {
	return &Server{
		router:  router,
		address: address,
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.address)
}
