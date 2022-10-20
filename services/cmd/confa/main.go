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

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/cmd/confa/queue"
	"github.com/aromancev/confa/cmd/confa/web"
	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/confa/talk/clap"
	"github.com/aromancev/confa/internal/proto/rtc"
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
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()
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

	rtcClient := rtc.NewRTCProtobufClient("http://"+config.RTCRPCAddress, &http.Client{})

	confaMongo := confa.NewMongo(mongoDB)
	confaCRUD := confa.NewCRUD(confaMongo)
	talkMongo := talk.NewMongo(mongoDB)
	talkBeanstalk := talk.NewBeanstalk(producer, talk.Tubes{
		StartRecording: config.Beanstalk.TubeStartRecording,
		StopRecording:  config.Beanstalk.TubeStopRecording,
	})
	talkUserService := talk.NewUserService(talkMongo, confaMongo, talkBeanstalk, rtcClient)
	clapMongo := clap.NewMongo(mongoDB)
	clapCRUD := clap.NewCRUD(clapMongo, talkMongo)
	profileEmitter := profile.NewBeanstalkEmitter(producer, profile.BeanstalkTubes{
		UpdateAvatar: config.Beanstalk.TubeUpdateAvatar,
	})
	profileMongo := profile.NewMongo(mongoDB)
	avatarUploader := profile.NewUpdater(
		config.Storage.PublicURL,
		profile.Buckets{
			UserUploads: config.Storage.BucketUserUploads,
			UserPublic:  config.Storage.BucketUserPublic,
		},
		minioClient,
		profileEmitter,
		profileMongo,
	)

	jobHandler := queue.NewHandler(
		avatarUploader,
		rtcClient,
		talkMongo,
		talkBeanstalk,
		queue.Tubes{
			UpdateAvatar:   config.Beanstalk.TubeUpdateAvatar,
			StartRecording: config.Beanstalk.TubeStartRecording,
			StopRecording:  config.Beanstalk.TubeStopRecording,
		},
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
				clapCRUD,
				profileMongo,
				avatarUploader,
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
