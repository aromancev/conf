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
	ListenWebAddress string `envconfig:"LISTEN_WEB_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	BaseURL          string `envconfig:"BASE_URL"`
	SecretKey        string `envconfig:"SECRET_KEY"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	Email            EmailConfig
	Mongo            MongoConfig
	Beanstalk        BeanstalkConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}

	sk, err := base64.StdEncoding.DecodeString(c.SecretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode SECRET_KEY (expected base64)")
	}
	c.SecretKey = string(sk)
	pk, err := base64.StdEncoding.DecodeString(c.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode PUBLIC_KEY (expected base64)")
	}
	c.PublicKey = string(pk)

	c.Email = c.Email.WithEnv()
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.ListenWebAddress == "" {
		return errors.New("ADDRESS not set")
	}
	if c.BaseURL == "" {
		return errors.New("BASE_URL not set")
	}
	if c.SecretKey == "" {
		return errors.New("SECRET_KEY not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	if err := c.Email.Validate(); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}
	if err := c.Mongo.Validate(); err != nil {
		return fmt.Errorf("invalid mongo config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
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
		return errors.New("iMONGO_USER not set")
	}
	if c.Password == "" {
		return errors.New("MONGO_PASSWORD not set")
	}
	if c.Database == "" {
		return errors.New("MONGO_DATABASE not set")
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
