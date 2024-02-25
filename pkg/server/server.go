package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i any) error {
	return v.validator.Struct(i)
}

type Status string

const (
	StatusNone      Status = ""
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type HostStatus struct {
	Status     Status `json:"status"`
	LastStatus Status `json:"-"`

	ReportedAt time.Time `json:"reported_at"`
}

type Server struct {
	config   *config.ServerConfig
	echoInst *echo.Echo

	statuses sync.Map
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.monitorStatuses(ctx)

	return s.echoInst.Start(s.config.ListenAddress)
}

func (s *Server) monitorStatuses(ctx context.Context) {
	ticker := time.NewTicker(s.config.UpdateFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Stopping status monitor")
			return
		case <-ticker.C:
			log.Debug().Msg("Updating statuses")

			s.statuses.Range(func(key any, value any) bool {
				identifier := key.(string)
				status := value.(*HostStatus)

				if time.Since(status.ReportedAt) >= s.config.UnhealthyTime {
					status.Status = StatusUnhealthy
				}

				if status.Status == status.LastStatus {
					return true
				}

				log.Info().
					Str("identifier", identifier).
					Str("last_status", string(status.LastStatus)).
					Str("status", string(status.Status)).
					Msg("Host status changed")

				status.LastStatus = status.Status
				return true
			})
		}
	}
}

func (s *Server) getStatuses(c echo.Context) error {
	statuses := map[string]*HostStatus{}

	s.statuses.Range(func(key any, value any) bool {
		statuses[key.(string)] = value.(*HostStatus)
		return true
	})

	return c.JSON(http.StatusOK, statuses)
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

	existingStatus, hasStatus := s.statuses.Load(body.Identifier)
	if !hasStatus {
		s.statuses.Store(body.Identifier, &HostStatus{ReportedAt: time.Now(), Status: StatusHealthy})
		return c.JSON(http.StatusOK, map[string]any{})
	}

	status := existingStatus.(*HostStatus)
	status.ReportedAt = time.Now()
	status.Status = StatusHealthy

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

	echoInst.GET("/statuses", serverInst.getStatuses)
	echoInst.POST("/ping", serverInst.requireAuth(serverInst.handlePing))

	return serverInst
}
