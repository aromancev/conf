package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aromancev/confa/cmd/avp/queue"
	"github.com/aromancev/confa/internal/dash"
	s3cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config := Config{}.WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		log.Logger = zerolog.New(os.Stdout)
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()
	switch config.LogLevel {
	case LevelDebug:
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case LevelWarn:
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}
	ctx = log.Logger.WithContext(ctx)

	producer, err := beanstalk.NewProducer(config.Beanstalk.ParsePool(), beanstalk.Config{
		Multiply:         1,
		ReconnectTimeout: 3 * time.Second,
		InfoFunc: func(message string) {
			log.Info().Msg(message)
		},
		ErrorFunc: func(err error, message string) {
			log.Err(err).Msg(message)
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to beanstalk.")
	}
	consumer, err := beanstalk.NewConsumer(config.Beanstalk.ParsePool(), []string{config.Beanstalk.TubeProcessTrack}, beanstalk.Config{
		Multiply:         1,
		NumGoroutines:    1,
		ReserveTimeout:   1 * time.Second,
		ReconnectTimeout: 3 * time.Second,
		InfoFunc: func(message string) {
			log.Info().Msg(message)
		},
		ErrorFunc: func(err error, message string) {
			log.Err(err).Msg(message)
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to beanstalk.")
	}

	cfg, err := s3cfg.LoadDefaultConfig(ctx, s3cfg.WithRegion(config.Storage.Region))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create s3 config.")
	}
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		endpoint := fmt.Sprintf("%s://%s", config.Storage.Scheme, config.Storage.Host)
		o.BaseEndpoint = &endpoint
		o.Credentials = credentials.NewStaticCredentialsProvider(config.Storage.AccessKey, config.Storage.SecretKey, "")
		o.UsePathStyle = true
	})

	tubes := queue.Tubes{
		ProcessTrack:         config.Beanstalk.TubeProcessTrack,
		UpdateRecordingTrack: config.Beanstalk.TubeUpdateRecordingTrack,
	}
	jobHandler := queue.NewHandler(
		dash.NewConverter(s3Client, config.Storage.BucketTrackPublic),
		tubes,
		queue.NewBeanstalk(producer, tubes),
	)

	var servers sync.WaitGroup
	servers.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		servers.Done()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()
	cancel()

	servers.Wait()
	producer.Stop()

	log.Info().Msg("Shutdown complete")
}
