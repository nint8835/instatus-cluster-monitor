package server

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

type Server struct {
	config *config.ServerConfig

	echoInst *echo.Echo
}

func (s *Server) Start() error {
	return s.echoInst.Start(s.config.ListenAddress)
}

func (s *Server) handlePing(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{})
}

func New(c *config.ServerConfig) *Server {
	echoInst := echo.New()
	serverInst := &Server{
		config:   c,
		echoInst: echoInst,
	}

	echoInst.HideBanner = true

	echoInst.POST("/ping", serverInst.requireAuth(serverInst.handlePing))

	return serverInst
}
