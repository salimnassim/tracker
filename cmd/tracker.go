package main

import (
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	state := make(chan any)
	errors := make(chan error)
	stop := make(chan bool)

	server := tracker.NewServer(errors, stop, state)
	udp := tracker.NewUDPServer("0.0.0.0", 8888)

	go server.Start(udp)

	for {
		select {
		case <-stop:
			return
		case err := <-errors:
			if err != nil {
				log.Error().Err(err)
			}
		case event := <-state:
			log.Debug().Msgf("%v", event)
		}

	}
}
