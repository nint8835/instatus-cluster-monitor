package config

import (
	"fmt"
)

type ServerConfig struct {
	SharedConfig

	// ListenAddress is the address to listen on
	ListenAddress string `split_words:"true" default:":8080"`
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
