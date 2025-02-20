package main

import (
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	state := make(chan any)

	conns := tracker.NewStore[uint64, uint64]()
	torrents := tracker.NewStore[[20]byte, *tracker.Torrent]()

	server := tracker.NewServer(state, conns, torrents)
	udp := tracker.NewUDPServer("127.0.0.1", 8888)

	go server.Start(udp)
	log.Info().Msg("started")

	for event := range state {
		switch e := event.(type) {
		case tracker.EventConnection:
			log.Info().Uint64("connection_id", e.ConnectionID).Msg("new connection")
			conns.Set(e.ConnectionID, 0)
		case error:
			log.Error().Err(e)
		}
	}
}
