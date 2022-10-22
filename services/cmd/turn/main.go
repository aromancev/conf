package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/aromancev/confa/auth"
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
	switch config.LogLevel {
	case LevelDebug:
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}

	tcpListener, err := net.Listen("tcp4", config.ListenWebAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen on socket.")
	}

	publicKey, err := auth.NewPublicKey(config.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create public key")
	}

	s, err := turn.NewServer(turn.ServerConfig{
		LoggerFactory: LoggerFactory(func(scope string) logging.LeveledLogger {
			return NewLogger(log.Logger)
		}),
		Realm: config.Realm,
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			// Passing jwt via username because I don't know how to customise auth otherwise.
			var claims auth.APIClaims
			if err := publicKey.Verify(username, &claims); err != nil {
				return nil, false
			}
			return turn.GenerateAuthKey(username, config.Realm, config.Realm), true
		},
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

	log.Info().Msg("TURN listening on " + config.ListenWebAddress)

	// Block until user sends SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	_ = s.Close()
}
