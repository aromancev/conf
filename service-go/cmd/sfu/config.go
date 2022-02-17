package main

import (
	"errors"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"
)

type Config struct {
	Address       string `envconfig:"ADDRESS"`
	ICEPortMin    uint16 `envconfig:"ICE_PORT_MIN"`
	ICEPortMax    uint16 `envconfig:"ICE_PORT_MAX"`
	ICEUrls       string `envconfig:"ICE_URLS"`
	ICEUsername   string `envconfig:"ICE_USERNAME"`
	ICECredential string `envconfig:"ICE_CREDENTIAL"`
	LogFormat     string `envconfig:"LOG_FORMAT"`
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	return c
}

func (c Config) WithDefault() Config {
	c.Address = ":8080"
	c.ICEPortMin = 5000
	c.ICEPortMax = 5200
	return c
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address not set")
	}
	if c.ICEPortMin == 0 {
		return errors.New("ice port min not set")
	}
	if c.ICEPortMax == 0 {
		return errors.New("ice port max not set")
	}

	return nil
}
