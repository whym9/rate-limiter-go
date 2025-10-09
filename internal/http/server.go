package http

import (
	"fmt"
	log "rate-limiter-go/internal/log"

	"github.com/labstack/echo/v4"
)

type Server struct {
	logger *log.Logger
	addr   string
	router *echo.Echo
}

func New(logger *log.Logger, addr string) *Server {
	s := Server{
		logger: logger,
		addr:   addr,
	}
	s.router = echo.New()
	return &s
}

func (s *Server) Init() error {
	if err := s.router.Start(s.addr); err != nil {
		s.logger.Error(fmt.Sprintf("could not start server on adrees: %s", s.addr), err)
		return err
	}
	s.logger.Log(fmt.Sprintf("server started on address: %s", s.addr))
	return nil
}

func (s *Server) Router() *echo.Echo {
	return s.router
}
