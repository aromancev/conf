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

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/room"
	"github.com/aromancev/confa/room/record"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/cmd/rtc/queue"
	"github.com/aromancev/confa/cmd/rtc/rpc"
	"github.com/aromancev/confa/cmd/rtc/web"
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

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		config.Mongo.User,
		config.Mongo.Password,
		config.Mongo.Hosts,
		config.Mongo.Database,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongo")
	}
	mongoDB := mongoClient.Database(config.Mongo.Database)

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
		log.Fatal().Err(err).Msg("Failed to connect to beanstalkd")
	}
	consumer, err := beanstalk.NewConsumer(config.Beanstalk.ParsePool(), []string{config.Beanstalk.TubeStoreEvent}, beanstalk.Config{
		Multiply:         1,
		NumGoroutines:    10,
		ReserveTimeout:   100 * time.Millisecond,
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

	sfuConn, err := grpc.DialContext(ctx, config.RTC.SFUAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to sfu RPC.")
	}

	publicKey, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	roomMongo := room.NewMongo(mongoDB)
	recordMongo := record.NewMongo(mongoDB)
	eventMongo := event.NewMongo(mongoDB)
	eventWatcher := event.NewSharedWatcher(eventMongo, 30)
	eventEmitter := event.NewBeanstalkEmitter(producer, config.Beanstalk.TubeStoreEvent)

	jobHandler := queue.NewHandler(eventMongo, queue.Tubes{
		StoreEvent: config.Beanstalk.TubeStoreEvent,
	})

	webServer := &http.Server{
		Addr:         config.ListenWebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(
			publicKey,
			roomMongo,
			eventMongo,
			event.NewBeanstalkEmitter(producer, config.Beanstalk.TubeStoreEvent),
			sfuConn,
			eventWatcher,
		),
	}
	rpcServer := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:         config.ListenRPCAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: rtc.NewRTCServer(
			rpc.NewHandler(
				roomMongo,
				recordMongo,
				tracker.NewRegistryProtobufClient(config.TrackerRPCAddress, &http.Client{}),
				eventEmitter,
			),
		),
	}

	go func() {
		log.Info().Msg("Web listening on " + config.ListenWebAddress)
		if err := webServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Web server failed")
		}
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

	go func() {
		log.Info().Msg("Serving event watcher.")
		if err := eventWatcher.Serve(ctx, 60*time.Second); err != nil {
			if errors.Is(err, event.ErrShuttingDown) {
				return
			}
			log.Fatal().Err(err).Msg("Event watcher failed.")
		}
	}()

	var consumerDone sync.WaitGroup
	consumerDone.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		consumerDone.Done()
	}()
	log.Info().Msg("Listening on " + config.ListenWebAddress)

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
	_ = sfuConn.Close()

	log.Info().Msg("Shutdown complete")
}
