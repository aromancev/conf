package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/logging"
	"github.com/pion/turn/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	// Create a TCP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any TCP listeners, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	tcpListener, err := net.Listen("tcp4", config.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen on socket.")
	}

	key := turn.GenerateAuthKey(config.Username, config.Realm, config.Credential)

	s, err := turn.NewServer(turn.ServerConfig{
		LoggerFactory: LoggerFactory(func(scope string) logging.LeveledLogger {
			return NewLogger(log.Logger)
		}),
		Realm: config.Realm,
		// Set AuthHandler callback
		// This is called everytime a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			return key, true
		},
		// ListenerConfig is a list of Listeners and the configuration around them
		ListenerConfigs: []turn.ListenerConfig{
			{
				Listener: tcpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(config.PublicIP),
					Address:      "0.0.0.0",
				},
			},
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TURN server.")
	}

	log.Info().Msg("TURN listening on " + config.Address)

	// Block until user sends SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	_ = s.Close()
}
