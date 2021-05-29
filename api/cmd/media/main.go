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

	avp "github.com/pion/ion-avp/pkg"
	"github.com/prep/beanstalk"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/aromancev/confa/cmd/media/handler"
	"github.com/aromancev/confa/internal/media/video"
	"github.com/aromancev/confa/proto/media"
	"github.com/aromancev/confa/proto/queue"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config := Config{}.WithDefault().WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config.")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()

	producer, err := beanstalk.NewProducer(config.Beanstalkd.Pool, beanstalk.Config{
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
		log.Fatal().Err(err).Msg("Failed to connect to beanstalkd")
	}
	consumer, err := beanstalk.NewConsumer(config.Beanstalkd.Pool, []string{queue.TubeVideo}, beanstalk.Config{
		Multiply:         1,
		NumGoroutines:    10,
		ReserveTimeout:   5 * time.Second,
		ReconnectTimeout: 3 * time.Second,
		InfoFunc: func(message string) {
			log.Info().Msg(message)
		},
		ErrorFunc: func(err error, message string) {
			log.Err(err).Msg(message)
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to beanstalkd")
	}

	grpcServer := grpc.NewServer()

	avpConf := avp.Config{
		SampleBuilder: avp.Samplebuilderconf{
			AudioMaxLate:  100,
			VideoMaxLate:  200,
			MaxLateTimeMs: 1000,
		},
	}
	avpConf.WebRTC.PLICycle = 1000
	sfuFactory := grpcpool.Factory(func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(config.SFUAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Err(err).Msg("Failed to start gRPC connection.")
			return nil, err
		}
		log.Info().Msg("SFU GRPC connection opened.")
		return conn, nil
	})
	sfuPool, err := grpcpool.New(sfuFactory, 0, 3, 10*time.Second)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create verifier")
	}
	media.RegisterAVPServer(grpcServer, handler.NewAVP(avpConf, sfuPool))

	videoConverter := video.NewConverter(config.MediaDir)

	mediaHandler := http.FileServer(http.Dir(config.MediaDir))
	httpHandler := handler.NewHTTP(mediaHandler)
	jobHandler := handler.NewJob(videoConverter)

	handler.InitAVP(config.MediaDir, producer)

	srv := &http.Server{
		Addr:         config.HTTPAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      httpHandler,
	}
	go func() {
		log.Info().Msg("HTTP Listening on " + config.HTTPAddress)
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	var consumerDone sync.WaitGroup
	consumerDone.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		consumerDone.Done()
	}()

	go func() {
		lis, err := net.Listen("tcp", config.RPCAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to listen TCP.")
		}
		log.Info().Msg("RPC listening on " + config.RPCAddress)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve GRPC.")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()

	cancel()
	_ = srv.Shutdown(ctx)
	grpcServer.Stop()
	producer.Stop()
	consumerDone.Wait()

	log.Info().Msg("Shutdown complete")
}
