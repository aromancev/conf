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
	Address       string `envconfig:"ADDRESS"`
	LogFormat     string `envconfig:"LOG_FORMAT"`
	BaseURL       string `envconfig:"BASE_URL"`
	SecretKey     string `envconfig:"SECRET_KEY"`
	PublicKey     string `envconfig:"PUBLIC_KEY"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR"`
	Email         EmailConfig
	Postgres      PostgresConfig
	Beanstalkd    BeanstalkdConfig
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

	c.Email, err = c.Email.Parsed()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse email config")
	}
	c.Beanstalkd, err = c.Beanstalkd.Parsed()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse beanstalkd config")
	}
	return c
}

func (c Config) WithDefault() Config {
	c.Address = ":80"
	return c
}

func (c Config) Validate() error {
	if c.Address == "" {
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
	if c.MigrationsDir == "" {
		return errors.New("MIGRATIONS_DIR not set")
	}
	if err := c.Email.Validate(); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}
	if err := c.Postgres.Validate(); err != nil {
		return fmt.Errorf("invalid postgres config: %w", err)
	}
	if err := c.Beanstalkd.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalkd config: %w", err)
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

func (c EmailConfig) Parsed() (EmailConfig, error) {
	pass, err := base64.StdEncoding.DecodeString(c.Password)
	if err != nil {
		return EmailConfig{}, errors.New("failed to password (expected base64)")
	}
	c.Password = string(pass)
	return c, nil
}

type BeanstalkdConfig struct {
	RawPool string `envconfig:"BEANSTALKD_POOL"`
	Pool    []string
}

func (c BeanstalkdConfig) Validate() error {
	if c.RawPool == "" {
		return errors.New("pool not set")
	}

	return nil
}

func (c BeanstalkdConfig) Parsed() (BeanstalkdConfig, error) {
	c.Pool = strings.Split(c.RawPool, ",")
	return c, nil
}
