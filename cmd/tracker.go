package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	ctx := context.Background()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// create config
	config := tracker.NewServerConfig(
		os.Getenv("ADDRESS"),
		os.Getenv("ANNOUNCE_URL"),
		os.Getenv("DSN"),
		os.Getenv("TEMPLATE_PATH"),
	)

	// create server
	server := tracker.NewServer(config)

	// create router
	r := mux.NewRouter()
	r.Handle("/health", tracker.HealthHandler())

	r.Handle("/", tracker.IndexHandler(server))
	r.Handle("/torrent/{id}", tracker.TorrentHandler(server))

	r.Handle("/announce", tracker.AnnounceHandler(server))

	log.Info().Msgf("starting tracker (address: %s, announce url: %s)", config.Address, config.AnnounceURL)

	http := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      5 * time.Second,
		Addr:              config.Address,
		Handler:           r,
	}

	// remove stale peers every 5 minutes
	server.RunTask(5*time.Minute, func(ts tracker.TorrentStorable) {
		_, err := ts.CleanPeers(ctx, 1*time.Hour*24)
		if err != nil {
			log.Error().Err(err).Msg("cant clean peers in task")
			return
		}
		// todo: prom
	})

	err := http.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("tracker exited")
	}
}
