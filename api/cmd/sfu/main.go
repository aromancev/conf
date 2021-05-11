package main

import (
	"net"
	"os"
	"time"

	avp "github.com/pion/ion-avp/pkg"
	ilog "github.com/pion/ion-log"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/aromancev/confa/cmd/sfu/server"
	pavp "github.com/aromancev/confa/proto/avp"
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
		},
	})

	dc := sfuServer.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	avpConf := avp.Config{
		SampleBuilder: avp.Samplebuilderconf{
			AudioMaxLate: 100,
			VideoMaxLate: 200,
		},
	}
	ilog.Init("debug", nil, nil)
	avpConf.Log.Level = "debug"
	avpConf.WebRTC.PLICycle = 1000
	sfuFactory := grpcpool.Factory(func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(config.Address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Err(err).Msg("Failed to start gRPC connection.")
			return nil, err
		}
		log.Info().Msg("SFU GRPC connection opened.")
		return conn, nil
	})
	sfuPool, err := grpcpool.New(sfuFactory, 0, 3, 10*time.Second)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create verifier")
	}

	psfu.RegisterSFUServer(grpcServer, server.NewSFU(sfuServer))
	pavp.RegisterAVPServer(grpcServer, server.NewAVP(sfuPool, avpConf))

	lis, err := net.Listen("tcp", config.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen TCP.")
	}
	log.Info().Msg("Listening on " + config.Address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve GRPC.")
	}
}
