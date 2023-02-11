package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aromancev/confa/cmd/sender-clients/web"
	"github.com/aromancev/confa/cmd/sender-clients/web/mailersend"
	"github.com/rs/zerolog/log"
)

func main() {
	config := Config{}.WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	webServer := &http.Server{
		Addr:         config.ListenWebAddress,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      web.NewHandler(mailersend.NewMailersend()),
	}

	go func() {
		log.Info().Msg("Web listening on " + config.ListenWebAddress)
		if err := webServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("Web server failed")
		}
	}()

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt)
	defer done()

	<-ctx.Done()

	log.Info().Msg("Shutting down.")
	ctx, shutdown := context.WithTimeout(context.Background(), time.Second*60)
	defer shutdown()

	_ = webServer.Shutdown(ctx)
	log.Info().Msg("Shutdown complete.")
}
