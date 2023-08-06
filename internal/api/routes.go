package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (s *Server) initRoutes() {
	s.GET("/swagger/*", echoSwagger.WrapHandler)

	s.GET("/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	v1 := s.Group("/api/v1")
	{
		//TODO revisit
		fmt.Println(v1)
	}
}
