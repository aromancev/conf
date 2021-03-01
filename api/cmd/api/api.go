package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/cmd/api/handler"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/psql"
)

func main() {
	ctx := context.Background()

	config := Config{}.WithDefault().WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()

	postgres, err := psql.New(ctx, psql.Config{
		Host:     config.Postgres.Host,
		Port:     config.Postgres.Port,
		User:     config.Postgres.User,
		Password: config.Postgres.Password,
		Database: config.Postgres.Database,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgres")
	}

	confaSQL := confa.NewSQL()
	confaCRUD := confa.NewCRUD(postgres, confaSQL)

	srv := &http.Server{
		Addr:         config.Address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler.New(confaCRUD),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()
	log.Info().Msg("Listening on " + config.Address)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	_ = srv.Shutdown(ctx)

	log.Fatal().Msg("Shutting down")
}
