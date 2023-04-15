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
	ListenRPCAddress string `envconfig:"LISTEN_RPC_ADDRESS"`
	WebHost          string `envconfig:"WEB_HOST"`
	WebScheme        string `envconfig:"WEB_SCHEME"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	SecretKey        string `envconfig:"SECRET_KEY"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	Mongo            MongoConfig
	Beanstalk        BeanstalkConfig
	Google           Google
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
	if c.WebHost == "" {
		return errors.New("WEB_HOST not set")
	}
	if c.WebScheme == "" {
		return errors.New("WEB_SCHEME not set")
	}
	if c.SecretKey == "" {
		return errors.New("SECRET_KEY not set")
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
	if err := c.Google.Validate(); err != nil {
		return fmt.Errorf("invalid google config: %w", err)
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

type BeanstalkConfig struct {
	Pool             string `envconfig:"BEANSTALK_POOL"`
	TubeSend         string `envconfig:"BEANSTALK_TUBE_SEND"`
	TubeUpdateAvatar string `envconfig:"BEANSTALK_TUBE_UPDATE_AVATAR"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeSend == "" {
		return errors.New("BEANSTALK_TUBE_SEND not set")
	}
	if c.TubeUpdateAvatar == "" {
		return errors.New("BEANSTALK_TUBE_UPDATE_AVATAR not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type Google struct {
	APIBaseURL   string `envconfig:"GOOGLE_API_BASE_URL"`
	ClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	ClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
}

func (c Google) Validate() error {
	if c.APIBaseURL == "" {
		return errors.New("GOOGLE_API_BASE_URL not set")
	}
	if c.ClientID == "" {
		return errors.New("GOOGLE_CLIENT_ID not set")
	}
	if c.ClientSecret == "" {
		return errors.New("GOOGLE_CLIENT_SECRET not set")
	}
	return nil
}
