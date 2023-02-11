package main

import (
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
	MailersendBaseURL   string `envconfig:"MAILERSEND_BASE_URL"`
	MailersendToken     string `envconfig:"MAILERSEND_TOKEN"`
	MailersendFromEmail string `envconfig:"MAILERSEND_FROM_EMAIL"`
}

func (c EmailConfig) Validate() error {
	if c.MailersendBaseURL == "" {
		return errors.New("MAILERSEND_BASE_URL not set")
	}
	if c.MailersendToken == "" {
		return errors.New("MAILERSEND_TOKEN not set")
	}
	if c.MailersendFromEmail == "" {
		return errors.New("MAILERSEND_FROM_EMAIL not set")
	}
	return nil
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
