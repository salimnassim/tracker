package main

import (
	"encoding/hex"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	state := make(chan any)

	conns := tracker.NewStore[uint64, uint64]()
	torrents := tracker.NewStore[[20]byte, *tracker.Torrent]()

	server := tracker.NewServer(state, conns, torrents)
	udp := tracker.NewUDPServer("", 8888)

	go server.Start(udp)
	log.Info().Msg("started")

	for event := range state {
		switch e := event.(type) {
		case tracker.EventConnection:
			log.Info().Uint64("connection_id", e.ConnectionID).Msg("connection")
			conns.Set(e.ConnectionID, 0)
		case tracker.EventAnnounce:
			log.Info().Str("peer_id", hex.EncodeToString(e.PeerId[:])).Msg("announce")
		case tracker.EventRegisterTorrent:
			log.Info().Str("info_hash", hex.EncodeToString(e.InfoHash[:])).Msg("register torrent")

			torrent := tracker.NewTorrent()
			torrents.Set(e.InfoHash, torrent)
		case error:
			log.Error().Err(e)
		}
	}
}
