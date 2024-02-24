package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") != fmt.Sprintf("Bearer %s", s.config.SharedSecret) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing Authorization header")
		}

		return next(c)
	}
}
