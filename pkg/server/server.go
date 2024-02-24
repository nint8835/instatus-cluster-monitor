package server

import (
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i any) error {
	return v.validator.Struct(i)
}

type Server struct {
	config   *config.ServerConfig
	echoInst *echo.Echo

	statuses sync.Map
}

func (s *Server) Start() error {
	return s.echoInst.Start(s.config.ListenAddress)
}

type PingBody struct {
	Identifier string `json:"identifier" validate:"required"`
}

func (s *Server) handlePing(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") != "application/json" {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}

	var body PingBody

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{})
}

func New(c *config.ServerConfig) *Server {
	echoInst := echo.New()
	serverInst := &Server{
		config:   c,
		echoInst: echoInst,
	}

	echoInst.HideBanner = true
	echoInst.Validator = &Validator{validator: validator.New()}

	echoInst.POST("/ping", serverInst.requireAuth(serverInst.handlePing))

	return serverInst
}
