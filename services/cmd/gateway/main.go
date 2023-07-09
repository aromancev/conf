package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/movio/bramble"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	logrus "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

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
		// Bramble uses global logrus logger, so we have to change log level for it manually.
		logrus.SetLevel(logrus.DebugLevel)
	case LevelWarn:
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
		logrus.SetLevel(logrus.WarnLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
		logrus.SetLevel(logrus.InfoLevel)
	}
	ctx = log.Logger.WithContext(ctx)

	plugins := []bramble.Plugin{
		&PassHeaders{},
		&TraceHeader{},
	}
	var services []*bramble.Service
	for _, s := range config.ParseServices() {
		services = append(services, bramble.NewService(s))
	}
	gateway := bramble.NewGateway(
		bramble.NewExecutableSchema(
			plugins,
			10,
			bramble.NewClientWithPlugins(plugins),
			services...,
		),
		plugins,
	)

	go gateway.UpdateSchemas(time.Duration(config.SchemaUpdateIntervalS) * time.Second)

	server := &http.Server{
		Addr:         config.ListenWebAddress,
		Handler:      gateway.Router(&bramble.Config{}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server terminated unexpectedly.")
	}

	<-ctx.Done()
	cancel()
	log.Info().Msg("Sutting down.")

	ctx, cancelShutdown := context.WithTimeout(ctx, 5*time.Second)
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to shut down server.")
	}
	cancelShutdown()
}
