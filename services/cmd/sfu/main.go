package main

import (
	"net"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pion/ion-sfu/cmd/signal/grpc/server"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/ion/proto/rtc"
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

	sfu.Logger = logr.New(NewLogger(log.Logger))

	grpcServer := grpc.NewServer()
	sfuConfig := sfu.Config{
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
			ICEPortRange: []uint16{config.ICEPortMin, config.ICEPortMax},
			SDPSemantics: "unified-plan",
			MDNS:         true,
			Timeouts: sfu.WebRTCTimeoutsConfig{
				ICEDisconnectedTimeout: 5,
				ICEFailedTimeout:       25,
				ICEKeepaliveInterval:   2,
			},
		},
	}
	if config.ICEUrls != "" {
		sfuConfig.WebRTC.ICEServers = append(sfuConfig.WebRTC.ICEServers, sfu.ICEServerConfig{
			URLs:       strings.Split(config.ICEUrls, ","),
			Username:   config.ICEUsername,
			Credential: config.ICECredential,
		})
	}
	sfuService := sfu.NewSFU(sfuConfig)

	dc := sfuService.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	rtc.RegisterRTCServer(grpcServer, server.NewSFUServer(sfuService))

	lis, err := net.Listen("tcp", config.ListenWebAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen TCP.")
	}
	log.Info().Msg("Listening on " + config.ListenWebAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve GRPC.")
	}
}
