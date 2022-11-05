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
	RTCRPCAddress    string `envconfig:"RTC_RPC_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	Mongo            MongoConfig
	Storage          StorageConfig
	Beanstalk        BeanstalkConfig
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
	if c.ListenWebAddress == "" {
		return errors.New("LISTEN_WEB_ADDRESS not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	if c.RTCRPCAddress == "" {
		return errors.New("RTC_RPC_ADDRESS not set")
	}
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if err := c.Mongo.Validate(); err != nil {
		return fmt.Errorf("invalid mongo config: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage config: %w", err)
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

type StorageConfig struct {
	Host              string `envconfig:"STORAGE_HOST"`
	AccessKey         string `envconfig:"STORAGE_ACCESS_KEY"`
	SecretKey         string `envconfig:"STORAGE_SECRET_KEY"`
	PublicURL         string `envconfig:"STORAGE_PUBLIC_URL"`
	BucketUserUploads string `envconfig:"STORAGE_BUCKET_USER_UPLOADS"`
	BucketUserPublic  string `envconfig:"STORAGE_BUCKET_USER_PUBLIC"`
}

func (c StorageConfig) Validate() error {
	if c.Host == "" {
		return errors.New("STORAGE_HOST not set")
	}
	if c.AccessKey == "" {
		return errors.New("STORAGE_ACCESS_KEY not set")
	}
	if c.SecretKey == "" {
		return errors.New("STORAGE_SECRET_KEY not set")
	}
	if c.PublicURL == "" {
		return errors.New("STORAGE_PUBLIC_URL not set")
	}
	if c.BucketUserUploads == "" {
		return errors.New("STORAGE_BUCKET_USER_UPLOADS not set")
	}
	if c.BucketUserPublic == "" {
		return errors.New("STORAGE_BUCKET_USER_PUBLIC not set")
	}
	return nil
}

type BeanstalkConfig struct {
	Pool               string `envconfig:"BEANSTALK_POOL"`
	TubeUpdateAvatar   string `envconfig:"BEANSTALK_TUBE_UPDATE_AVATAR"`
	TubeStartRecording string `envconfig:"BEANSTALK_TUBE_START_RECORDING"`
	TubeStopRecording  string `envconfig:"BEANSTALK_TUBE_STOP_RECORDING"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeUpdateAvatar == "" {
		return errors.New("BEANSTALK_TUBE_UPDATE_AVATAR not set")
	}
	if c.TubeStartRecording == "" {
		return errors.New("BEANSTALK_TUBE_START_RECORDING not set")
	}
	if c.TubeStopRecording == "" {
		return errors.New("BEANSTALK_TUBE_STOP_RECORDING not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}
