package api

import (
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"

	"github.com/levongh/profile/internal/config"
	"github.com/levongh/profile/internal/log"
)

type Server struct {
	*echo.Echo

	cfg     *config.Config
	Logger  *log.Logger
	handler Handler
	// ss      *ServiceStorage

	closeJaeger io.Closer
}

type Handler struct {
	db     *sqlx.DB
	logger *log.Logger
}

func NewServer(cfg *config.Config, logger *log.Logger) (*Server, error) {
	var err error

	s := &Server{
		Echo:   echo.New(),
		cfg:    cfg,
		Logger: logger,
	}

	s.handler = Handler{
		db:     nil,
		logger: logger,
	}

	s.initRoutes()
	// s.initMidleware()

	s.closeJaeger = jaegertracing.New(s.Echo, nil)
	return s, err //TODO revisit
}

// func (s *Server) ServiceStorage() *ServiceStorage {
// return s.ss
// }

func (s *Server) Close() error {
	var allErrors error

	if err := s.Echo.Close(); err != nil {
		allErrors = addError(allErrors, err)
	}

	// if err := s.ss.Close(); err != nil {
	// 	allErrors = addError(allErrors, err)
	// }

	if err := s.closeJaeger.Close(); err != nil {
		allErrors = addError(allErrors, err)
	}

	return allErrors
}

func addError(allErrors, err error) error {
	if allErrors == nil {
		return err
	}
	return fmt.Errorf("%s; %s", allErrors.Error(), err.Error())
}
