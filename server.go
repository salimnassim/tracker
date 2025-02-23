package tracker

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/rs/zerolog/log"
)

type eventConnection struct {
	ConnectionID uint64
}

type eventAnnounce struct {
	InfoHash   InfoHash
	PeerId     PeerID
	Downloaded uint64
	Left       uint64
	Uploaded   uint64
	Event      uint32
	IP         uint32
	Key        uint32
	Port       uint16
}

type eventRegisterTorrent struct {
	InfoHash [20]byte
}

type Serverer interface {
	Serve(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent])
	Address() string
	Port() int
}

type server struct {
	ctx      context.Context
	conns    Storer[uint64, uint64]
	torrents Storer[InfoHash, *Torrent]

	state chan any
}

func NewServer(conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) *server {
	return &server{
		ctx:      context.Background(),
		conns:    conns,
		torrents: torrents,
		state:    make(chan any),
	}
}

func (s *server) Start(servers []Serverer) error {
	for _, server := range servers {
		go server.Serve(s.state, s.conns, s.torrents)
		log.Info().Str("address", server.Address()).Int("port", server.Port()).Msg("started server")
	}

	for event := range s.state {
		switch e := event.(type) {
		case eventConnection:
			s.conns.Set(e.ConnectionID, uint64(time.Now().Unix()))

			log.Info().Uint64("connection_id", e.ConnectionID).Msg("connection created")
		case eventAnnounce:
			log.Info().Str("peer_id", hex.EncodeToString(e.PeerId[:])).Msg("announce")

			torrent, ok := s.torrents.Get(e.InfoHash)
			if !ok {
				log.Error().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("announce torrent not found")
				continue
			}

			peer, ok := torrent.Peers[e.PeerId]
			if !ok {
				torrent.Peers[e.PeerId] = &Peer{
					Event:      e.Event,
					Left:       e.Left,
					Downloaded: e.Downloaded,
					Uploaded:   e.Uploaded,
					IP:         e.IP,
					Port:       e.Port,
					Key:        e.Key,
					Time:       time.Now().Unix(),
				}

				log.Info().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("announce peer created")
				continue
			}

			if peer.Key != 0 && peer.Key != e.Key {
				log.Info().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("announce peer key mismatch")
				continue
			}

			if e.Event == 1 {
				torrent.Completed = torrent.Completed + 1

				log.Info().
					Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
					Str("peer_id", hex.EncodeToString(e.PeerId[:])).
					Msg("announce completed")
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

		case eventRegisterTorrent:
			torrent := NewTorrent(e.InfoHash)
			s.torrents.Set(e.InfoHash, torrent)

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Msg("torrent registered")

		case error:
			log.Error().Err(e)
		}
	}

	return nil
}
