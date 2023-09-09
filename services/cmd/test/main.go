package main

import (
	"context"
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	var roomID string
	flag.StringVar(&roomID, "r", "", "room to join")

	flag.Parse()

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger.Level(zerolog.TraceLevel)

	ctx = log.Logger.WithContext(ctx)

	NewTracker(ctx, LivekitCredentials{
		URL:    "wss://sfu.confa.io",
		Key:    "key",
		Secret: "93d33a06-f209-4239-bd7f-d04d411ae7b2",
	}, roomID)

	select {}
}
