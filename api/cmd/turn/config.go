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
	Address    string `envconfig:"ADDRESS"`
	Realm      string `envconfig:"REALM"`
	LogFormat  string `envconfig:"LOG_FORMAT"`
	Username   string `envconfig:"USERNAME"`
	Credential string `envconfig:"CREDENTIAL"`
	PublicIP   string `envconfig:"PUBLIC_IP"`
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	return c
}

func (c Config) WithDefault() Config {
	return c
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address not set")
	}
	if c.Realm == "" {
		return errors.New("realm not set")
	}
	if c.Username == "" {
		return errors.New("username not set")
	}
	if c.Credential == "" {
		return errors.New("credential not set")
	}
	if c.PublicIP == "" {
		return errors.New("public IP not set")
	}
	return nil
}
