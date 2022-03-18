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
	LogFormat        string `envconfig:"LOG_FORMAT"`
	BaseURL          string `envconfig:"BASE_URL"`
	SecretKey        string `envconfig:"SECRET_KEY"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	Email            EmailConfig
	Mongo            MongoConfig
	Beanstalkd       BeanstalkdConfig
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
	if err := c.Beanstalkd.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalkd config: %w", err)
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

type EmailConfig struct {
	Server   string `envconfig:"EMAIL_SERVER"`
	Port     string `envconfig:"EMAIL_PORT"`
	Address  string `envconfig:"EMAIL_ADDRESS"`
	Password string `envconfig:"EMAIL_PASSWORD"`
	Secure   string `envconfig:"EMAIL_SECURE"`
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

func (c EmailConfig) WithEnv() EmailConfig {
	pass, err := base64.StdEncoding.DecodeString(c.Password)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode email password")
	}
	c.Password = string(pass)
	return c
}

type BeanstalkdConfig struct {
	Pool          string `envconfig:"BEANSTALKD_POOL"`
	TubeSendEmail string `envconfig:"BEANSTALKD_TUBE_SEND_EMAIL"`
}

func (c BeanstalkdConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("pool not set")
	}
	if c.TubeSendEmail == "" {
		return errors.New("tube 'send email' not set")
	}

	return nil
}

func (c BeanstalkdConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}
