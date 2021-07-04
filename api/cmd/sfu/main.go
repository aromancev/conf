package main

import (
	"net"
	"os"

	psfu "github.com/pion/ion-sfu/cmd/signal/grpc/proto"
	"github.com/pion/ion-sfu/cmd/signal/grpc/server"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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

	// TODO: Implement interface
	// sfu.Logger = log.Logger

	grpcServer := grpc.NewServer()
	sfuService := sfu.NewSFU(sfu.Config{
		Router: sfu.RouterConfig{
			MaxBandwidth:        1500,
			MaxPacketTrack:      500,
			AudioLevelThreshold: 40,
			AudioLevelInterval:  1000,
			AudioLevelFilter:    20,
			Simulcast: sfu.SimulcastConfig{
				BestQualityFirst: true,
			},
		},
		WebRTC: sfu.WebRTCConfig{
			ICEServers: []sfu.ICEServerConfig{
				{
					URLs: []string{"stun:stun.stunprotocol.org:3478", "stun:stun.l.google.com:19302"},
				},
			},
			ICEPortRange: []uint16{config.ICEPortMin, config.ICEPortMax},
			SDPSemantics: "unified-plan",
			MDNS:         true,
		},
	})

	dc := sfuService.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	psfu.RegisterSFUServer(grpcServer, server.NewServer(sfuService))

	lis, err := net.Listen("tcp", config.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen TCP.")
	}
	log.Info().Msg("Listening on " + config.Address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve GRPC.")
	}
}
