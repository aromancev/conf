package main

import (
	"net"
	"os"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/aromancev/confa/cmd/sfu/handler"
	psfu "github.com/aromancev/confa/proto/sfu"
)

func main() {
	config := Config{}.WithDefault().WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config.")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()

	grpcServer := grpc.NewServer()
	sfuServer := sfu.NewSFU(sfu.Config{
		WebRTC: sfu.WebRTCConfig{
			ICEPortRange: []uint16{config.ICEPortMin, config.ICEPortMax},
		},
	})
	dc := sfuServer.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(handler.SubscriberAPI)

	psfu.RegisterSFUServer(grpcServer, handler.NewServer(sfuServer))

	lis, err := net.Listen("tcp", config.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen TCP.")
	}
	log.Info().Msg("Listening on " + config.Address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve GRPC.")
	}
}
