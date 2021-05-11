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
	Address   string `envconfig:"ADDRESS"`
	LogFormat string `envconfig:"LOG_FORMAT"`
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
	return c
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address not set")
	}

	return nil
}
