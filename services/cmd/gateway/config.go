package main

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"

	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelWarn  = "warn"
)

type Config struct {
	ListenWebAddress      string `envconfig:"LISTEN_WEB_ADDRESS"`
	LogFormat             string `envconfig:"LOG_FORMAT"`
	LogLevel              string `envconfig:"LOG_LEVEL"`
	Services              string `envconfig:"SERVICES"`
	SchemaUpdateIntervalS int    `envconfig:"SCHEMA_UPDATE_INTERVAL_S"`
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.ListenWebAddress == "" {
		return errors.New("LISTEN_WEB_ADDRESS not set")
	}
	if c.Services == "" {
		return errors.New("SERVICES not set")
	}
	if c.SchemaUpdateIntervalS == 0 {
		return errors.New("SCHEMA_UPDATE_INTERVAL_S not set")
	}
	return nil
}

func (c Config) ParseServices() []string {
	return strings.Split(c.Services, ",")
}
