package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/cmd/confa/web"
	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/confa/talk/clap"
	"github.com/aromancev/confa/internal/proto/rtc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	publicKey, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	rtcClient := rtc.NewRTCProtobufClient(config.RTCAddress, &http.Client{})

	confaMongo := confa.NewMongo(mongoDB)
	confaCRUD := confa.NewCRUD(confaMongo)
	talkMongo := talk.NewMongo(mongoDB)
	talkCRUD := talk.NewCRUD(talkMongo, confaMongo, rtcClient)
	clapMongo := clap.NewMongo(mongoDB)
	clapCRUD := clap.NewCRUD(clapMongo, talkMongo)

	webServer := &http.Server{
		Addr:         config.ListenWebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(web.NewResolver(
			publicKey,
			confaCRUD,
			talkCRUD,
			clapCRUD,
		)),
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()
	cancel()

	_ = webServer.Shutdown(ctx)
	log.Info().Msg("Shutdown complete")
}
