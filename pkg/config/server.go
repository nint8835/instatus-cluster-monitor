package config

import (
	"fmt"
	"time"
)

type ServerConfig struct {
	SharedConfig

	// ListenAddress is the address to listen on
	ListenAddress string `split_words:"true" default:":8080"`

	// UnhealthyTime is the amount of time since last ping before a host is marked unhealthy
	UnhealthyTime time.Duration `split_words:"true" default:"5m"`

	// UpdateFrequency is the amount of time between status updates
	UpdateFrequency time.Duration `split_words:"true" default:"1m"`
}

func LoadServerConfig() (*ServerConfig, error) {
	config, err := loadConfig[ServerConfig]()
	if err != nil {
		return nil, err
	}

	err = initLoggingConfig(config.SharedConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing logging: %w", err)
	}

	return config, nil
}
