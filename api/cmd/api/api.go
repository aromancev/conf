package main

import (
	"context"
	"errors"
	"github.com/aromancev/confa/internal/confa/talk"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/ident"
	"github.com/aromancev/confa/internal/user/session"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/api/handler"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/psql"
	"github.com/aromancev/confa/internal/platform/trace"
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

	pgConf := psql.Config{
		Host:     config.Postgres.Host,
		Port:     config.Postgres.Port,
		User:     config.Postgres.User,
		Password: config.Postgres.Password,
		Database: config.Postgres.Database,
	}
	postgres, err := psql.New(ctx, pgConf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgres")
	}

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
	consumer, err := beanstalk.NewConsumer(config.Beanstalkd.Pool, []string{handler.TubeEmail}, beanstalk.Config{
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

	sign, err := auth.NewSigner(config.SecretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create signer")
	}
	verify, err := auth.NewVerifier(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create verifier")
	}

	sender := email.NewSender(config.Email.Server, config.Email.Port, config.Email.Address, config.Email.Password, config.Email.Secure != "false")

	confaSQL := confa.NewSQL()
	confaCRUD := confa.NewCRUD(postgres, confaSQL)

	talkSQL := talk.NewSQL()
	talkCRUD := talk.NewCRUD(postgres, talkSQL, confaCRUD)

	sessionSQL := session.NewSQL()
	sessionCRUD := session.NewCRUD(postgres, sessionSQL)

	userSQL := user.NewSQL()

	identSQL := ident.NewSQL()
	identCRUD := ident.NewCRUD(postgres, identSQL, userSQL)

	hand := handler.New(config.BaseURL, confaCRUD, talkCRUD, sessionCRUD, identCRUD, sender, trace.NewBeanstalkd(producer), sign, verify)

	srv := &http.Server{
		Addr:         config.Address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      hand,
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
		consumer.Receive(ctx, hand.ServeJob)
		consumerDone.Done()
	}()
	log.Info().Msg("Listening on " + config.Address)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Fatal().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()

	cancel()
	_ = srv.Shutdown(ctx)
	producer.Stop()
	consumerDone.Wait()

	log.Fatal().Msg("Shutdown complete")
}
