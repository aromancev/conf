package main

import (
	"encoding/base64"
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
	ListenWebAddress string `envconfig:"LISTEN_WEB_ADDRESS"`
	Realm            string `envconfig:"REALM"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	Username         string `envconfig:"USERNAME"`
	Credential       string `envconfig:"CREDENTIAL"`
	PublicIP         string `envconfig:"PUBLIC_IP"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	pk, err := base64.StdEncoding.DecodeString(c.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode PUBLIC_KEY (expected base64)")
	}
	c.PublicKey = string(pk)
	return c
}

func (c Config) WithDefault() Config {
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.ListenWebAddress == "" {
		return errors.New("LISTEN_WEB_ADDRESS not set")
	}
	if c.Realm == "" {
		return errors.New("REALM not set")
	}
	if c.Username == "" {
		return errors.New("USERNAME not set")
	}
	if c.Credential == "" {
		return errors.New("CREDENTIAL not set")
	}
	if c.PublicIP == "" {
		return errors.New("PUBLIC_IP not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	return nil
}
