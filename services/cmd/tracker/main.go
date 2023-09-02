package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	pb "github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/tracker/record"
	s3cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/tracker/rpc"
	"github.com/aromancev/confa/tracker"
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

	runtime := tracker.NewRuntime()

	rpcServer := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:         config.ListenRPCAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: pb.NewRegistryServer(
			rpc.NewHandler(
				runtime,
				s3Client,
				record.NewBeanstalk(producer, record.Tubes{
					ProcessTrack:         config.Beanstalk.TubeProcessTrack,
					StoreEvent:           config.Beanstalk.TubeStoreEvent,
					UpdateRecordingTrack: config.Beanstalk.TubeUpdateRecordingTrack,
				}),
				rpc.Buckets{
					TrackRecords: config.Storage.BucketTrackRecords,
				},
				record.LivekitCredentials{
					URL:    config.Livekit.URL,
					Key:    config.Livekit.Key,
					Secret: config.Livekit.Secret,
				},
			),
		),
	}

	var servers sync.WaitGroup
	servers.Add(1)
	go func() {
		defer servers.Done()

		runtime.Run(ctx, 60*time.Second)
	}()

	go func() {
		log.Info().Msg("RPC listening on " + config.ListenRPCAddress)
		if err := rpcServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("RPC server failed.")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()
	cancel()

	_ = rpcServer.Shutdown(ctx)
	producer.Stop()
	servers.Wait()

	log.Info().Msg("Shutdown complete")
}
