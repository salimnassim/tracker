package main

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	errs := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// server := tracker.NewServer()

	udp := tracker.NewUDPServer(
		"0.0.0.0", 8888, tracker.NewStore[uint64, uint32](), tracker.NewStore[[20]byte, *tracker.Torrent](),
	)

	go udp.Serve(errs)

	for {
		select {
		case <-stop:
			return
		case err := <-errs:
			if err != nil {
				log.Error().Err(err)
			}
		}
	}
}
