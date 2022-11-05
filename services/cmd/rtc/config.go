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
	ListenWebAddress  string `envconfig:"LISTEN_WEB_ADDRESS"`
	ListenRPCAddress  string `envconfig:"LISTEN_RPC_ADDRESS"`
	TrackerRPCAddress string `envconfig:"TRACKER_RPC_ADDRESS"`
	LogFormat         string `envconfig:"LOG_FORMAT"`
	LogLevel          string `envconfig:"LOG_LEVEL"`
	PublicKey         string `envconfig:"PUBLIC_KEY"`
	Mongo             MongoConfig
	Beanstalk         BeanstalkConfig
	RTC               RTCConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}

	pk, err := base64.StdEncoding.DecodeString(c.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode PUBLIC_KEY (expected base64)")
	}
	c.PublicKey = string(pk)
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
	if c.ListenRPCAddress == "" {
		return errors.New("LISTEN_RPC_ADDRESS not set")
	}
	if c.TrackerRPCAddress == "" {
		return errors.New("TRACKER_RPC_ADDRESS not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	if err := c.Mongo.Validate(); err != nil {
		return fmt.Errorf("invalid mongo config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
	}
	if err := c.RTC.Validate(); err != nil {
		return fmt.Errorf("invalid rtc config: %w", err)
	}

	return nil
}

type MongoConfig struct {
	Hosts    string `envconfig:"MONGO_HOSTS"`
	User     string `envconfig:"MONGO_USER"`
	Password string `envconfig:"MONGO_PASSWORD"`
	Database string `envconfig:"MONGO_DATABASE"`
}

func (c MongoConfig) Validate() error {
	if c.Hosts == "" {
		return errors.New("MONGO_HOSTS not set")
	}
	if c.User == "" {
		return errors.New("MONGO_USER not set")
	}
	if c.Password == "" {
		return errors.New("MONGO_PASSWORD not set")
	}
	if c.Database == "" {
		return errors.New("MONGO_DATABASE not set")
	}
	return nil
}

type BeanstalkConfig struct {
	Pool           string `envconfig:"BEANSTALK_POOL"`
	TubeStoreEvent string `envconfig:"BEANSTALK_TUBE_STORE_EVENT"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeStoreEvent == "" {
		return errors.New("BEANSTALK_TUBE_STORE_EVENT not set")
	}

	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type RTCConfig struct {
	SFURPCAddress string `envconfig:"SFU_RPC_ADDRESS"`
}

func (c RTCConfig) Validate() error {
	if c.SFURPCAddress == "" {
		return errors.New("SFU_RPC_ADDRESS not set")
	}
	return nil
}
