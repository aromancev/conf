package main

import (
	"errors"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"
)

type Config struct {
	Address   string `envconfig:"ADDRESS"`
	LogFormat string `envconfig:"LOG_FORMAT"`
	Email     EmailConfig
}

type EmailConfig struct {
	Server   string `envconfig:"EMAIL_SERVER"`
	Port     string `envconfig:"EMAIL_PORT"`
	Address  string `envconfig:"EMAIL_ADDRESS"`
	Password string `envconfig:"EMAIL_PASSWORD"`
}

func (c EmailConfig) Validate() error {
	if c.Server == "" {
		return errors.New("server not set")
	}
	if c.Port == "" {
		return errors.New("port not set")
	}
	if c.Address == "" {
		return errors.New("address not set")
	}
	if c.Password == "" {
		return errors.New("password not set")
	}
	return nil
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	return c
}

func (c Config) WithDefault() Config {
	c.Address = ":80"
	return c
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address not set")
	}
	if err := c.Email.Validate(); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}

	return nil
}
