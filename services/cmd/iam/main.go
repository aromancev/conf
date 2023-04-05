package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aromancev/confa/cmd/iam/rpc"
	"github.com/aromancev/confa/cmd/iam/web"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/platform/google/gsi"
	"github.com/aromancev/confa/internal/proto/iam"
	"github.com/aromancev/confa/internal/routes"
	"github.com/aromancev/confa/session"
	"github.com/aromancev/confa/user"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/prep/beanstalk"
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
	switch config.LogLevel {
	case LevelDebug:
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case LevelWarn:
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}
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
		log.Fatal().Err(err).Msg("Failed to connect to beanstalk.")
	}

	secretKey, err := auth.NewSecretKey(config.SecretKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create secret key")
	}
	publicKey, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	pages := routes.NewPages(config.WebScheme, config.WebHost)

	userMongo := user.NewMongo(mongoDB)
	userCRUD := user.NewActions(userMongo)
	sessionMongo := session.NewMongo(mongoDB)

	webServer := &http.Server{
		Addr:         config.ListenWebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(
			web.Tubes{
				Send:         config.Beanstalk.TubeSend,
				UpdateAvatar: config.Beanstalk.TubeUpdateAvatar,
			},
			web.NewAuth(config.WebHost),
			pages,
			secretKey,
			publicKey,
			web.NewResolver(
				publicKey,
				userMongo,
			),
			sessionMongo,
			userCRUD,
			producer,
			gsi.NewPublicKey(&http.Client{}),
			gsi.Creds{
				ClientID:     config.Google.ClientID,
				ClientSecret: config.Google.ClientSecret,
			},
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
		Handler: iam.NewIAMServer(
			rpc.NewHandler(userMongo),
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
}
