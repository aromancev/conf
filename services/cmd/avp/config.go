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
	LogFormat string `envconfig:"LOG_FORMAT"`
	LogLevel  string `envconfig:"LOG_LEVEL"`
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
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
	}
	return nil
}

type BeanstalkConfig struct {
	Pool                     string `envconfig:"BEANSTALK_POOL"`
	TubeProcessTrack         string `envconfig:"BEANSTALK_TUBE_PROCESS_TRACK"`
	TubeUpdateRecordingTrack string `envconfig:"BEANSTALK_TUBE_UPDATE_RECORDING_TRACK"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeProcessTrack == "" {
		return errors.New("BEANSTALK_TUBE_PROCESS_TRACK not set")
	}
	if c.TubeUpdateRecordingTrack == "" {
		return errors.New("BEANSTALK_TUBE_UPDATE_RECORDING_TRACK not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type StorageConfig struct {
	Host               string `envconfig:"STORAGE_HOST"`
	Scheme             string `envconfig:"STORAGE_SCHEME"`
	Region             string `envconfig:"STORAGE_REGION"`
	AccessKey          string `envconfig:"STORAGE_ACCESS_KEY"`
	SecretKey          string `envconfig:"STORAGE_SECRET_KEY"`
	BucketTrackRecords string `envconfig:"STORAGE_BUCKET_TRACK_RECORDS"`
	BucketTrackPublic  string `envconfig:"STORAGE_BUCKET_TRACK_PUBLIC"`
}

func (c StorageConfig) Validate() error {
	if c.Host == "" {
		return errors.New("STORAGE_HOST not set")
	}
	if c.Scheme == "" {
		return errors.New("STORAGE_SCHEME not set")
	}
	if c.Region == "" {
		return errors.New("STORAGE_REGION not set")
	}
	if c.AccessKey == "" {
		return errors.New("STORAGE_ACCESS_KEY not set")
	}
	if c.SecretKey == "" {
		return errors.New("STORAGE_SECRET_KEY not set")
	}
	if c.BucketTrackRecords == "" {
		return errors.New("STORAGE_BUCKET_TRACK_RECORDS not set")
	}
	if c.BucketTrackPublic == "" {
		return errors.New("STORAGE_BUCKET_TRACK_PUBLIC not set")
	}
	return nil
}
