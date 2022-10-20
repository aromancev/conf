package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aromancev/confa/event"
	pb "github.com/aromancev/confa/internal/proto/tracker"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/tracker/rpc"
	evtrack "github.com/aromancev/confa/event/tracker"
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
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()
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

	minioClient, err := minio.New(config.Storage.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Storage.AccessKey, config.Storage.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create minio client.")
	}

	connector := sdk.NewConnector(config.SFURPCAddress)
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
				connector,
				runtime,
				minioClient,
				evtrack.NewBeanstalk(producer, evtrack.Tubes{
					ProcessTrack: config.Beanstalk.TubeProcessTrack,
				}),
				event.NewBeanstalkEmitter(producer, config.Beanstalk.TubeStoreEvent),
				rpc.Buckets{
					TrackRecords: config.Storage.BucketTrackRecords,
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
	connector.Close()
	servers.Wait()

	log.Info().Msg("Shutdown complete")
}
