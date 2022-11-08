package main

import (
	"encoding/base64"
	"errors"
	"fmt"
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
	LogFormat     string `envconfig:"LOG_FORMAT"`
	LogLevel      string `envconfig:"LOG_LEVEL"`
	IAMRPCAddress string `envconfig:"IAM_RPC_ADDRESS"`
	Email         EmailConfig
	Beanstalk     BeanstalkConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}

	c.Email = c.Email.WithEnv()
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.IAMRPCAddress == "" {
		return errors.New("IAM_RPC_ADDRESS not set")
	}
	if err := c.Email.Validate(); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
	}
	return nil
}

type EmailConfig struct {
	Server   string `envconfig:"EMAIL_SERVER"`
	Port     string `envconfig:"EMAIL_PORT"`
	Address  string `envconfig:"EMAIL_ADDRESS"`
	Password string `envconfig:"EMAIL_PASSWORD"`
	Secure   string `envconfig:"EMAIL_SECURE"`
}

func (c EmailConfig) Validate() error {
	if c.Server == "" {
		return errors.New("EMAIL_SERVER not set")
	}
	if c.Port == "" {
		return errors.New("EMAIL_PORT not set")
	}
	if c.Address == "" {
		return errors.New("EMAIL_ADDRESS not set")
	}
	if c.Password == "" {
		return errors.New("EMAIL_PASSWORD not set")
	}
	return nil
}

func (c EmailConfig) WithEnv() EmailConfig {
	pass, err := base64.StdEncoding.DecodeString(c.Password)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode email password")
	}
	c.Password = string(pass)
	return c
}

type BeanstalkConfig struct {
	Pool     string `envconfig:"BEANSTALK_POOL"`
	TubeSend string `envconfig:"BEANSTALK_TUBE_SEND"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeSend == "" {
		return errors.New("BEANSTALK_TUBE_SEND not set")
	}

	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}
