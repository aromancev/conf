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

	"github.com/aromancev/confa/cmd/confa/queue"
	"github.com/aromancev/confa/cmd/confa/web"
	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/aromancev/confa/internal/routes"
	"github.com/aromancev/confa/profile"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prep/beanstalk"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config := Config{}.WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config.")
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
	case LevelWarn:
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}
	ctx = log.Logger.WithContext(ctx)

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
	consumer, err := beanstalk.NewConsumer(
		config.Beanstalk.ParsePool(),
		[]string{
			config.Beanstalk.TubeUpdateAvatar,
			config.Beanstalk.TubeStartRecording,
			config.Beanstalk.TubeStopRecording,
			config.Beanstalk.TubeRecordingUpdate,
		},
		beanstalk.Config{
			Multiply:         1,
			NumGoroutines:    3,
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
		log.Fatal().Err(err).Msg("Failed to connect to beanstalk.")
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		config.Mongo.User,
		config.Mongo.Password,
		config.Mongo.Hosts,
		config.Mongo.Database,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongo.")
	}
	mongoDB := mongoClient.Database(config.Mongo.Database)

	publicKey, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key.")
	}

	minioClient, err := minio.New(config.Storage.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Storage.AccessKey, config.Storage.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create minio client.")
	}

	pages := routes.NewPages(config.WebScheme, config.WebHost)
	storageRoutes := routes.NewStorage(config.Storage.PublicScheme, config.Storage.PublicDomain, routes.Buckets{
		UserPublic: config.Storage.BucketUserPublic,
	})

	rtcClient := rtc.NewRTCProtobufClient("http://"+config.RTCRPCAddress, &http.Client{})

	confaMongo := confa.NewMongo(mongoDB)
	confaCRUD := confa.NewUser(confaMongo)
	talkMongo := talk.NewMongo(mongoDB)
	talkBeanstalk := talk.NewBeanstalk(producer, talk.Tubes{
		StartRecording: config.Beanstalk.TubeStartRecording,
		StopRecording:  config.Beanstalk.TubeStopRecording,
	})
	talkUserService := talk.NewUser(talkMongo, confaMongo, talkBeanstalk, rtcClient)
	profileEmitter := profile.NewBeanstalkEmitter(producer, profile.BeanstalkTubes{
		UpdateAvatar: config.Beanstalk.TubeUpdateAvatar,
	})
	profileMongo := profile.NewMongo(mongoDB)
	avatarUploader := profile.NewUpdater(
		storageRoutes,
		profile.Buckets{
			UserUploads: config.Storage.BucketUserUploads,
			UserPublic:  config.Storage.BucketUserPublic,
		},
		minioClient,
		profileEmitter,
		profileMongo,
		&http.Client{},
	)

	tubes := queue.Tubes{
		UpdateAvatar:    config.Beanstalk.TubeUpdateAvatar,
		StartRecording:  config.Beanstalk.TubeStartRecording,
		StopRecording:   config.Beanstalk.TubeStopRecording,
		Send:            config.Beanstalk.TubeSend,
		RecordingUpdate: config.Beanstalk.TubeRecordingUpdate,
	}
	jobHandler := queue.NewHandler(
		avatarUploader,
		rtcClient,
		confaMongo,
		talkMongo,
		talkBeanstalk,
		tubes,
		queue.NewBeanstalk(producer, tubes),
		pages,
		profileMongo,
	)

	webServer := &http.Server{
		Addr:         config.ListenWebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: web.NewHandler(
			web.NewResolver(
				publicKey,
				confaCRUD,
				talkUserService,
				profileMongo,
				avatarUploader,
				storageRoutes,
			),
			publicKey,
		),
	}

	go func() {
		log.Info().Msg("Web listening on " + config.ListenWebAddress)
		if err := webServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Web server failed.")
		}
	}()

	var consumerDone sync.WaitGroup
	consumerDone.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		consumerDone.Done()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Shutting down.")

	ctx, shutdown := context.WithTimeout(ctx, time.Second*60)
	defer shutdown()
	cancel()

	_ = webServer.Shutdown(ctx)
	producer.Stop()
	consumerDone.Wait()
	log.Info().Msg("Shutdown complete.")
}
