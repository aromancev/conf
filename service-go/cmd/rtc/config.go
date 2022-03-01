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
)

type Config struct {
	ListenWebAddress string `envconfig:"LISTEN_WEB_ADDRESS"`
	ListenRPCAddress string `envconfig:"LISTEN_RPC_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	Mongo            MongoConfig
	Beanstalkd       BeanstalkdConfig
	RTC              RTCConfig
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

func (c Config) WithDefault() Config {
	c.ListenWebAddress = ":80"
	return c
}

func (c Config) Validate() error {
	if c.ListenWebAddress == "" {
		return errors.New("ADDRESS not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	if err := c.Mongo.Validate(); err != nil {
		return fmt.Errorf("invalid mongo config: %w", err)
	}
	if err := c.Beanstalkd.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalkd config: %w", err)
	}
	if err := c.RTC.Validate(); err != nil {
		return fmt.Errorf("invalid rtc config: %w", err)
	}

	return nil
}

type PostgresConfig struct {
	Host     string `envconfig:"POSTGRES_HOST"`
	Port     uint16 `envconfig:"POSTGRES_PORT"`
	User     string `envconfig:"POSTGRES_USER"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
	Database string `envconfig:"POSTGRES_DATABASE"`
}

func (c PostgresConfig) Validate() error {
	if c.Host == "" {
		return errors.New("host not set")
	}
	if c.Port == 0 {
		return errors.New("port not set")
	}
	if c.User == "" {
		return errors.New("user not set")
	}
	if c.Password == "" {
		return errors.New("password not set")
	}
	if c.Database == "" {
		return errors.New("database not set")
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
		return errors.New("hosts not set")
	}
	if c.User == "" {
		return errors.New("iam user not set")
	}
	if c.Password == "" {
		return errors.New("iam password not set")
	}
	if c.Database == "" {
		return errors.New("iam database not set")
	}
	return nil
}

type BeanstalkdConfig struct {
	Pool           string `envconfig:"BEANSTALKD_POOL"`
	TubeStoreEvent string `envconfig:"BEANSTALKD_TUBE_STORE_EVENT"`
}

func (c BeanstalkdConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("pool not set")
	}
	if c.TubeStoreEvent == "" {
		return errors.New("tube `store event` not set")
	}

	return nil
}

func (c BeanstalkdConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type RTCConfig struct {
	SFUAddress string `envconfig:"RTC_SFU_ADDRESS"`
}

func (c RTCConfig) Validate() error {
	if c.SFUAddress == "" {
		return errors.New("sfu address not set")
	}
	return nil
}
