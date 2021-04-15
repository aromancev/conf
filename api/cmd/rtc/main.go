package main

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/jsonrpc2"
	rpcws "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/aromancev/confa/cmd/rtc/handler"
)

func main() {
	config := Config{}.WithDefault().WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid config")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()

	s := sfu.NewSFU(sfu.Config{
		WebRTC: sfu.WebRTCConfig{
			ICEPortRange: []uint16{config.ICEPortMin, config.ICEPortMax},
		},
	})
	dc := s.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  config.ReadBuffer,
		WriteBufferSize: config.WriteBuffer,
	}

	http.Handle("/v1/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Err(err).Msg("failed to upgrade connection")
			return
		}
		defer c.Close()

		p := handler.NewJSONSignal(sfu.NewPeer(s))
		defer p.Close()

		jc := jsonrpc2.NewConn(r.Context(), rpcws.NewObjectStream(c), p)
		<-jc.DisconnectNotify()
	}))
	log.Info().Msg("Listening on " + config.Address)
	err := http.ListenAndServe(config.Address, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
}
