package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/webrtc/v3"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/media/queue"
	"github.com/aromancev/confa/cmd/media/rpc"
	"github.com/aromancev/confa/cmd/media/web"
	"github.com/aromancev/confa/internal/media/video"
	pmedia "github.com/aromancev/confa/proto/media"
	pqueue "github.com/aromancev/confa/proto/queue"
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
	ctx = log.Logger.WithContext(ctx)

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
	consumer, err := beanstalk.NewConsumer(config.Beanstalkd.Pool, []string{pqueue.TubeVideo}, beanstalk.Config{
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

	videoConverter := video.NewConverter(config.MediaDir)

	jobHandler := queue.NewHandler(videoConverter)

	webServer := &http.Server{
		Addr:         config.WebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(
			http.FileServer(
				http.Dir(config.MediaDir),
			),
		),
	}
	rpcServer := &http.Server{
		Addr:         config.RPCAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: pmedia.NewMediaServer(rpc.NewHandler(
			config.SFUAddress,
			config.MediaDir,
			sdk.NewEngine(
				sdk.Config{
					WebRTC: sdk.WebRTCTransportConfig{
						Configuration: webrtc.Configuration{
							ICEServers: []webrtc.ICEServer{
								{
									URLs: []string{"stun:stun.stunprotocol.org:3478", "stun:stun.l.google.com:19302"},
								},
							},
						},
					},
				},
			),
			producer,
		)),
	}

	go func() {
		log.Info().Msg("Web listening on " + config.WebAddress)
		if err := webServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()
	go func() {
		log.Info().Msg("RPC listening on " + config.RPCAddress)
		if err := rpcServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// TODO: On talk start.
	// go func() {
	// 	client := pmedia.NewMediaProtobufClient("http://localhost"+config.RPCAddress, &http.Client{})
	// 	_, err := client.SaveTracks(ctx, &pmedia.Session{
	// 		TraceId:   "main",
	// 		SessionId: "test session",
	// 	})
	// 	if err != nil {
	// 		log.Err(err).Msg("Failed to start saving tracks")
	// 	}
	// }()

	var consumerDone sync.WaitGroup
	consumerDone.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		consumerDone.Done()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()

	cancel()
	_ = webServer.Shutdown(ctx)
	_ = rpcServer.Shutdown(ctx)
	producer.Stop()
	consumerDone.Wait()

	log.Info().Msg("Shutdown complete")
}
