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
	"github.com/aromancev/confa/internal/platform/email"
	pqueue "github.com/aromancev/confa/proto/queue"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/api/queue"
	"github.com/aromancev/confa/cmd/api/web"
	"github.com/aromancev/confa/internal/auth"
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
	consumer, err := beanstalk.NewConsumer(config.Beanstalkd.Pool, []string{pqueue.TubeEmail}, beanstalk.Config{
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

	sign, err := auth.NewSecretKey(config.SecretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create secret key")
	}
	verify, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	confaMongo := confa.NewMongo(mongoDB)
	confaCRUD := confa.NewCRUD(confaMongo)
	talkMongo := talk.NewMongo(mongoDB)
	talkCRUD := talk.NewCRUD(talkMongo, confaMongo)
	clapMongo := clap.NewMongo(mongoDB)
	clapCRUD := clap.NewCRUD(clapMongo, talkMongo)

	userMongo := user.NewMongo(mongoDB)
	userCRUD := user.NewCRUD(userMongo)
	sessionMongo := session.NewMongo(mongoDB)
	sessionCRUD := session.NewCRUD(sessionMongo)

	jobHandler := queue.NewHandler(
		email.NewSender(config.Email.Server, config.Email.Port, config.Email.Address, config.Email.Password, config.Email.Secure != "false"),
	)

	srv := &http.Server{
		Addr:         config.Address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.New(web.NewResolver(
			config.BaseURL,
			sign,
			verify,
			producer,
			userCRUD,
			sessionCRUD,
			confaCRUD,
			talkCRUD,
			clapCRUD,
		)),
	}
	go func() {
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
	log.Info().Msg("Listening on " + config.Address)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()

	cancel()
	_ = srv.Shutdown(ctx)
	producer.Stop()
	consumerDone.Wait()

	log.Info().Msg("Shutdown complete")
}
