package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	instatus_go "github.com/nint8835/instatus-go"
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

var instatusStatusMap = map[Status]string{
	StatusHealthy:   "OPERATIONAL",
	StatusUnhealthy: "MAJOROUTAGE",
}

type HostStatus struct {
	Status Status `json:"status"`

	ReportedAt time.Time `json:"reported_at"`

	lastStatus  Status
	componentId string
}

type Server struct {
	config   *config.ServerConfig
	echoInst *echo.Echo

	instatusClient *instatus_go.Client
	instatusPageId string

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

				if status.Status == status.lastStatus {
					return true
				}

				log.Info().
					Str("identifier", identifier).
					Str("last_status", string(status.lastStatus)).
					Str("status", string(status.Status)).
					Msg("Host status changed")

				if status.componentId == "" {
					log.Debug().Str("identifier", identifier).Msg("Component ID not cached, fetching components")

					// TODO: Don't re-fetch on every component - rate limit is 30 requests / 5 minutes (6 requests per minute)
					components, err := s.instatusClient.GetComponents(instatus_go.GetComponentsRequest{
						PageId: s.instatusPageId,
					})
					if err != nil {
						log.Error().Err(err).Msg("Error getting components")
						return false
					}

					for _, component := range components {
						if component.Name == identifier {
							log.Debug().Str("identifier", identifier).Str("component_id", component.Id).Msg("Component ID found")
							status.componentId = component.Id
							break
						}
					}

					if status.componentId == "" {
						log.Info().Str("identifier", identifier).Msg("Component ID not found, creating component for host")
						// TODO: Create component
					}
				}

				currentComponent, err := s.instatusClient.GetComponent(instatus_go.GetComponentRequest{
					PageId:      s.instatusPageId,
					ComponentId: status.componentId,
				})
				if err != nil {
					log.Error().Err(err).Msg("Error getting current component status")
					return false
				}

				status.lastStatus = status.Status

				if currentComponent.Status == "UNDERMAINTENANCE" {
					log.Debug().Str("identifier", identifier).Msg("Component is under maintenance, skipping status update")
					return true
				}

				newStatus := instatusStatusMap[status.Status]

				_, err = s.instatusClient.UpdateComponent(instatus_go.UpdateComponentRequest{
					PageId:      s.instatusPageId,
					ComponentId: status.componentId,
					UpdatedFields: instatus_go.UpdateComponentFields{
						Status: &newStatus,
					},
				})
				if err != nil {
					log.Error().Err(err).Msg("Error updating component status")
					return false
				}

				log.Debug().Str("identifier", identifier).Str("status", string(status.Status)).Msg("Component status updated")

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

func New(c *config.ServerConfig) (*Server, error) {
	echoInst := echo.New()
	serverInst := &Server{
		config:         c,
		echoInst:       echoInst,
		instatusClient: instatus_go.New(c.InstatusKey),
	}

	pages, err := serverInst.instatusClient.GetPages(instatus_go.GetPagesRequest{})
	if err != nil {
		return nil, fmt.Errorf("error listing pages: %w", err)
	}

	for _, page := range pages {
		if page.Subdomain == c.TargetSubdomain {
			log.Debug().Str("subdomain", c.TargetSubdomain).Str("id", page.Id).Msg("Found target subdomain")
			serverInst.instatusPageId = page.Id
			break
		}
	}

	if serverInst.instatusPageId == "" {
		return nil, fmt.Errorf("target subdomain not found")
	}

	echoInst.HideBanner = true
	echoInst.Validator = &Validator{validator: validator.New()}

	echoInst.GET("/statuses", serverInst.getStatuses)
	echoInst.POST("/ping", serverInst.requireAuth(serverInst.handlePing))

	return serverInst, nil
}
