package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/aromancev/confa/internal/event"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/grpcpool"
	"github.com/aromancev/confa/internal/room"
	pqueue "github.com/aromancev/confa/proto/queue"
	"github.com/aromancev/confa/proto/rtc"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/cmd/api/queue"
	"github.com/aromancev/confa/cmd/api/rpc"
	"github.com/aromancev/confa/cmd/api/web"
	"github.com/aromancev/confa/internal/confa"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config := Config{}.WithDefault().WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()
	ctx = log.Logger.WithContext(ctx)

	iamMongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		config.Mongo.IAMUser,
		config.Mongo.IAMPassword,
		config.Mongo.Hosts,
		config.Mongo.IAMDatabase,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongo")
	}
	iamMongoDB := iamMongoClient.Database(config.Mongo.IAMDatabase)

	rtcMongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		config.Mongo.RTCUser,
		config.Mongo.RTCPassword,
		config.Mongo.Hosts,
		config.Mongo.RTCDatabase,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongo")
	}
	rtcMongoDB := rtcMongoClient.Database(config.Mongo.RTCDatabase)

	confaMongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		config.Mongo.ConfaUser,
		config.Mongo.ConfaPassword,
		config.Mongo.Hosts,
		config.Mongo.ConfaDatabase,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongo")
	}
	confaMongoDB := confaMongoClient.Database(config.Mongo.ConfaDatabase)

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
	consumer, err := beanstalk.NewConsumer(config.Beanstalkd.Pool, []string{pqueue.TubeEmail, pqueue.TubeEvent}, beanstalk.Config{
		Multiply:         1,
		NumGoroutines:    10,
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
		log.Fatal().Err(err).Msg("Failed to connect to beanstalkd")
	}

	sfuPool, err := grpcpool.New(
		grpcpool.Factory(func() (*grpc.ClientConn, error) {
			return grpc.DialContext(ctx, config.RTC.SFUAddress, grpc.WithInsecure())
		}),
		0, 3, 10*time.Minute,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initiate sfu RPC pool.")
	}

	sign, err := auth.NewSecretKey(config.SecretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create secret key")
	}
	verify, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	rtcClient := rtc.NewRTCProtobufClient(config.RTCAddress, &http.Client{})

	userMongo := user.NewMongo(iamMongoDB)
	userCRUD := user.NewCRUD(userMongo)
	sessionMongo := session.NewMongo(iamMongoDB)
	sessionCRUD := session.NewCRUD(sessionMongo)

	confaMongo := confa.NewMongo(confaMongoDB)
	confaCRUD := confa.NewCRUD(confaMongo)
	talkMongo := talk.NewMongo(confaMongoDB)
	talkCRUD := talk.NewCRUD(talkMongo, confaMongo, rtcClient)
	clapMongo := clap.NewMongo(confaMongoDB)
	clapCRUD := clap.NewCRUD(clapMongo, talkMongo)

	roomMongo := room.NewMongo(rtcMongoDB)
	eventMongo := event.NewMongo(rtcMongoDB)
	eventWatcher := event.NewSharedWatcher(eventMongo, 30)

	jobHandler := queue.NewHandler(
		email.NewSender(config.Email.Server, config.Email.Port, config.Email.Address, config.Email.Password, config.Email.Secure != "false"),
		eventMongo,
	)

	webServer := &http.Server{
		Addr:         config.Address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(web.NewResolver(
			config.BaseURL,
			sign,
			verify,
			producer,
			userCRUD,
			sessionCRUD,
			confaCRUD,
			talkCRUD,
			clapCRUD,
			roomMongo,
			&websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
				ReadBufferSize:  config.RTC.ReadBuffer,
				WriteBufferSize: config.RTC.WriteBuffer,
			},
			sfuPool,
			eventWatcher,
		)),
	}
	rpcServer := &http.Server{
		Addr:         config.RPCAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      rtc.NewRTCServer(rpc.NewHandler(roomMongo)),
	}

	go func() {
		log.Info().Msg("Web listening on " + config.Address)
		if err := webServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Web server failed")
		}
	}()

	go func() {
		log.Info().Msg("RPC listening on " + config.RPCAddress)
		if err := rpcServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("RPC server failed.")
		}
	}()

	go func() {
		log.Info().Msg("Serving event watcher.")
		if err := eventWatcher.Serve(ctx, 10*time.Second); err != nil {
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
	log.Info().Msg("Listening on " + config.Address)

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
