package main

import (
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config := tracker.NewServerConfig(
		os.Getenv("ADDRESS"),
		os.Getenv("URL"),
		os.Getenv("DSN"))

	server := tracker.NewServer(config)

	r := mux.NewRouter()
	r.Handle("/health", tracker.HealthHandler())
	r.Handle("/announce", tracker.AnnounceHandler(server))

	log.Info().Msgf("starting tracker (address: %s, url: %s)", config.Address, config.URL)

	var wg sync.WaitGroup
	wg.Add(1)
	go server.Run(r)
	wg.Wait()

	log.Info().Msgf("exited tracker")
}
