package main

import (
	"errors"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"

	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
)

type Config struct {
	ListenRPCAddress string `envconfig:"LISTEN_RPC_ADDRESS"`
	ICEPortMin       uint16 `envconfig:"ICE_PORT_MIN"`
	ICEPortMax       uint16 `envconfig:"ICE_PORT_MAX"`
	ICEUrls          string `envconfig:"ICE_URLS"`
	ICEUsername      string `envconfig:"ICE_USERNAME"`
	ICECredential    string `envconfig:"ICE_CREDENTIAL"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.ListenRPCAddress == "" {
		return errors.New("LISTEN_RPC_ADDRESS not set")
	}
	if c.ICEPortMin == 0 {
		return errors.New("ICE_PORT_MIN not set")
	}
	if c.ICEPortMax == 0 {
		return errors.New("ICE_PORT_MAX not set")
	}
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	return nil
}
