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
	RTCAddress       string `envconfig:"RTC_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
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
		return errors.New("address not set")
	}
	if c.PublicKey == "" {
		return errors.New("public key not set")
	}
	if c.RTCAddress == "" {
		return errors.New("rtc address not set")
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
		return errors.New("host not set")
	}
	if c.AccessKey == "" {
		return errors.New("access key not set")
	}
	if c.SecretKey == "" {
		return errors.New("secret key not set")
	}
	if c.PublicURL == "" {
		return errors.New("public url not set")
	}
	if c.BucketUserUploads == "" {
		return errors.New("bucket user uploads not set")
	}
	if c.BucketUserPublic == "" {
		return errors.New("bucket user public not set")
	}
	return nil
}

type BeanstalkConfig struct {
	Pool             string `envconfig:"BEANSTALKD_POOL"`
	TubeUpdateAvatar string `envconfig:"BEANSTALKD_TUBE_UPDATE_AVATAR"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("pool not set")
	}
	if c.TubeUpdateAvatar == "" {
		return errors.New("tube `update avatar` not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}
