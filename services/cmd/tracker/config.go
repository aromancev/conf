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
	ListenRPCAddress string `envconfig:"LISTEN_RPC_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	SFURPCAddress    string `envconfig:"SFU_RPC_ADDRESS"`
	Beanstalk        BeanstalkConfig
	Storage          StorageConfig
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
	if c.ListenRPCAddress == "" {
		return errors.New("listen rpc address not set")
	}
	if c.SFURPCAddress == "" {
		return errors.New("sfu address not set")
	}
	if c.PublicKey == "" {
		return errors.New("public key not set")
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
	Pool             string `envconfig:"BEANSTALK_POOL"`
	TubeProcessTrack string `envconfig:"BEANSTALK_TUBE_PROCESS_TRACK"`
	TubeStoreEvent   string `envconfig:"BEANSTALK_TUBE_STORE_EVENT"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("pool not set")
	}
	if c.TubeProcessTrack == "" {
		return errors.New("tube `process track` not set")
	}
	if c.TubeStoreEvent == "" {
		return errors.New("tube `store event` not set")
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
	return nil
}
