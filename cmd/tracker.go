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
			conns.Set(e.ConnectionID, 0)

			log.Info().Uint64("connection_id", e.ConnectionID).Msg("connection created")
		case tracker.EventAnnounce:
			log.Info().Str("peer_id", hex.EncodeToString(e.PeerId[:])).Msg("announce")

			torrent, ok := torrents.Get(e.InfoHash)
			if !ok {
				log.Error().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("tried to announce non-existent torrent")
				continue
			}

			peer, ok := torrent.Peers[e.PeerId]
			if !ok {
				torrent.Peers[e.PeerId] = &tracker.Peer{
					Event:      e.Event,
					Left:       e.Left,
					Downloaded: e.Downloaded,
					Uploaded:   e.Uploaded,
					IP:         e.IP,
					Port:       e.Port,
				}

				log.Info().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("peer created")
				continue
			}

			peer.Event = e.Event
			peer.Left = e.Left
			peer.Downloaded = e.Downloaded
			peer.Uploaded = e.Uploaded
			peer.IP = e.IP
			peer.Port = e.Port

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Str("peer_id", hex.EncodeToString(e.PeerId[:])).
				Msg("peer updated")

		case tracker.EventRegisterTorrent:
			torrent := tracker.NewTorrent()
			torrents.Set(e.InfoHash, torrent)

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Msg("torrent registered")

		case error:
			log.Error().Err(e)
		}
	}
}
