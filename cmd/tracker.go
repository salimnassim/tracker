package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// create config
	config := tracker.NewServerConfig(
		os.Getenv("ADDRESS"),
		os.Getenv("URL"),
		os.Getenv("DSN"))

	// create server
	server := tracker.NewServer(config)

	// create router
	r := mux.NewRouter()
	r.Handle("/health", tracker.HealthHandler())
	r.Handle("/announce", tracker.AnnounceHandler(server))

	log.Info().Msgf("starting tracker (address: %s, url: %s)", config.Address, config.URL)

	// start goroutines
	http := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      10 * time.Second,
		Addr:              config.Address,
		Handler:           r,
	}

	err := http.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("tracker exited")
	}
}
