package config

import (
	"fmt"
	"time"
)

type AgentConfig struct {
	SharedConfig

	// PingFrequency is the amount of time between pings
	PingFrequency time.Duration `split_words:"true" default:"1m"`

	// ServerAddress is the address of the server to ping
	ServerAddress string `split_words:"true" required:"true"`

	// HostIdentifier is the identifier for this host.
	// If omitted, the host's hostname will be used.
	HostIdentifier string `split_words:"true"`
}

func LoadAgentConfig() (*AgentConfig, error) {
	config, err := loadConfig[AgentConfig]()
	if err != nil {
		return nil, err
	}

	err = initLoggingConfig(config.SharedConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing logging: %w", err)
	}

	return config, nil
}
