package main

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"
)

type Config struct {
	WebAddress string `envconfig:"WEB_ADDRESS"`
	RPCAddress string `envconfig:"RPC_ADDRESS"`
	SFUAddress string `envconfig:"SFU_ADDRESS"`
	LogFormat  string `envconfig:"LOG_FORMAT"`
	MediaDir   string `envconfig:"MEDIA_DIR"`

	Beanstalkd BeanstalkdConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
	}
	c.Beanstalkd, err = c.Beanstalkd.Parsed()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse beanstalkd config")
	}
	return c
}

func (c Config) WithDefault() Config {
	c.WebAddress = ":80"
	c.RPCAddress = ":8080"
	return c
}

func (c Config) Validate() error {
	if c.WebAddress == "" {
		return errors.New("http address not set")
	}
	if c.RPCAddress == "" {
		return errors.New("rpc address not set")
	}
	if c.SFUAddress == "" {
		return errors.New("SFU address not set")
	}
	if c.MediaDir == "" {
		return errors.New("media dir not set")
	}
	if err := c.Beanstalkd.Validate(); err != nil {
		return err
	}

	return nil
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
