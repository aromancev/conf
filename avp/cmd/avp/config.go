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
)

type Config struct {
	LogFormat string `envconfig:"LOG_FORMAT"`
	Beanstalk BeanstalkConfig
	Storage   StorageConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}
	return c
}

func (c Config) Validate() error {
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
	}
	return nil
}

type BeanstalkConfig struct {
	Pool             string `envconfig:"BEANSTALK_POOL"`
	TubeProcessTrack string `envconfig:"BEANSTALK_TUBE_PROCESS_TRACK"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("pool not set")
	}
	if c.TubeProcessTrack == "" {
		return errors.New("tube `process track` not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type StorageConfig struct {
	Host               string `envconfig:"STORAGE_HOST"`
	AccessKey          string `envconfig:"STORAGE_ACCESS_KEY"`
	SecretKey          string `envconfig:"STORAGE_SECRET_KEY"`
	BucketTrackRecords string `envconfig:"STORAGE_BUCKET_TRACK_RECORDS"`
	BucketTrackPublic  string `envconfig:"STORAGE_BUCKET_TRACK_PUBLIC"`
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
	if c.BucketTrackRecords == "" {
		return errors.New("bucket track records not set")
	}
	if c.BucketTrackPublic == "" {
		return errors.New("bucket track public not set")
	}
	return nil
}
