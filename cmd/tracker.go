package main

import (
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	state := make(chan any)

	conns := tracker.NewStore[uint64, uint64]()
	torrents := tracker.NewStore[tracker.InfoHash, *tracker.Torrent]()

	server := tracker.NewServer(state, conns, torrents)

	udp_port, err := strconv.Atoi(os.Getenv("BT_UDP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse udp port")
	}
	http_port, err := strconv.Atoi(os.Getenv("BT_HTTP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse http port")
	}

	udp := tracker.NewUDPServer("", udp_port)
	http := tracker.NewHTTPServer("", http_port)

	go server.StartUDP(udp)
	log.Info().Int("port", udp_port).Msg("started udp server")

	go server.StartHTTP(http)
	log.Info().Int("port", http_port).Msg("started http server")

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
					Time:       time.Now().Unix(),
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
			peer.Time = time.Now().Unix()

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Str("peer_id", hex.EncodeToString(e.PeerId[:])).
				Msg("peer updated")

		case tracker.EventRegisterTorrent:
			torrent := tracker.NewTorrent(e.InfoHash)
			torrents.Set(e.InfoHash, torrent)

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Msg("torrent registered")

		case error:
			log.Error().Err(e)
		}
	}
}
