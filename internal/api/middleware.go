package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	common "github.com/levongh/profile/common/config"
	"github.com/levongh/profile/common/httpx"
	// _ "github.com/levongh/profile/cmd/docs" // nolint:golint
)

func skipLoggingFunc(c echo.Context) bool {
	uri := c.Request().RequestURI
	return strings.Contains(uri, "health-check")
}

func (s *Server) makeIPCMiddleware(username, password string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, pass, ok := c.Request().BasicAuth()
			isUnAuthorized := !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1

			if isUnAuthorized {
				return c.JSON(http.StatusUnauthorized, nil)
			}
			return next(c)
		}
	}
}

func (s *Server) apiGatewayAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	if s.cfg.Mode == common.ModeLocal {
		return func(c echo.Context) error {
			// c.Set(string(httpx.ContextKeyUserID), s.cfg.MockUserID)
			return next(c)
		}
	}
	return httpx.APIGateWayAuthMiddleware(next)
}
