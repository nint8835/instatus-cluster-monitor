package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
	"github.com/nint8835/instatus-cluster-monitor/pkg/server"
)

type Agent struct {
	config *config.AgentConfig

	stopChan chan struct{}
}

func (a *Agent) run() {
	ticker := time.NewTicker(a.config.PingFrequency)
	defer ticker.Stop()

	log.Info().Msg("Starting agent")

	for {
		select {
		case <-a.stopChan:
			log.Debug().Msg("Stopping agent")
			return
		case <-ticker.C:
			log.Debug().Msg("Pinging")

			reqBody := server.PingBody{
				Identifier: a.config.HostIdentifier,
			}

			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				log.Error().Err(err).Msg("Error marshalling body")
				continue
			}

			req, err := http.NewRequest(http.MethodPost, a.config.ServerAddress, bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+a.config.SharedConfig.SharedSecret)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				log.Error().Err(err).Msg("Error creating request")
				continue
			}

			_, err = http.DefaultClient.Do(req)
			if err != nil {
				log.Error().Err(err).Msg("Error sending request")
				continue
			}
		}
	}
}

func (a *Agent) Stop() {
	a.stopChan <- struct{}{}
}

func (a *Agent) Start() {
	go a.run()
}

func New(c *config.AgentConfig) (*Agent, error) {
	if c.HostIdentifier == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("error getting hostname: %w", err)
		}
		c.HostIdentifier = hostname

		log.Debug().Str("identifier", c.HostIdentifier).Msg("Using hostname as identifier")
	}

	return &Agent{
		config:   c,
		stopChan: make(chan struct{}),
	}, nil
}
