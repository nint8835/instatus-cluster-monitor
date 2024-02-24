package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SharedConfig struct {
	// LogLevel is the log level to use
	LogLevel string `split_words:"true" default:"info"`

	// SharedSecret is the secret to use to authenticate with the server
	SharedSecret string `split_words:"true" required:"true"`
}

func initLoggingConfig(config SharedConfig) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("error parsing log level: %w", err)
	}
	zerolog.SetGlobalLevel(logLevel)

	return nil
}

func loadConfig[T any]() (*T, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error().Err(err).Msg("Failed to load .env file")
	}

	var config T

	err = envconfig.Process("instatus_monitor", &config)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	return &config, nil
}
