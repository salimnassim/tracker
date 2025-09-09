package tracker

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	errorServerArg = errors.New("server arg is nil")
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

type eventPeerLifetime struct{}

type Serverer interface {
	Serve(config *config, state chan any, conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent])
}

type server struct {
	ctx   context.Context
	state chan any

	conns Storer[uint64, time.Time]

	torrents Storer[InfoHash, *Torrent]
}

func NewServer(conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent]) (*server, error) {
	if conns == nil {
		return nil, fmt.Errorf("%w: conns store", errorServerArg)
	}
	if torrents == nil {
		return nil, fmt.Errorf("%w: torrents store", errorServerArg)
	}

	return &server{
		ctx:      context.Background(),
		conns:    conns,
		torrents: torrents,
		state:    make(chan any),
	}, nil
}

func (s *server) Start(config *config, servers []Serverer) error {
	for _, server := range servers {
		go server.Serve(config, s.state, s.conns, s.torrents)
		log.Info().Str("type", fmt.Sprintf("%T", server)).Msg("started server")
	}

	go func(s *server) {
		log.Info().Msg("started peer ticker")
		peerTicker := time.NewTicker(config.peerInterval * time.Second)
		for range peerTicker.C {
			s.state <- eventPeerLifetime{}
		}
	}(s)

	for event := range s.state {
		switch e := event.(type) {
		case eventConnection:
			s.conns.Set(e.ConnectionID, time.Now().UTC())

			log.Info().
				Uint64("connection_id", e.ConnectionID).
				Msg("connection created")
		case eventAnnounce:
			log.Info().
				Str("peer_id", hex.EncodeToString(e.PeerId[:])).
				Msg("announce")

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
					Time:       time.Now().UTC(),
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
			peer.Time = time.Now().UTC()

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Str("peer_id", hex.EncodeToString(e.PeerId[:])).
				Msg("peer updated")

		case eventRegisterTorrent:
			torrent := NewTorrent(e.InfoHash, config)
			s.torrents.Set(e.InfoHash, torrent)

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Msg("torrent registered")

		case eventPeerLifetime:
			s.torrents.Map(func(ih InfoHash, t *Torrent) {
				expiredPeers := []PeerID{}
				for peerID, peer := range t.Peers {
					if time.Now().After(peer.Time.Add(config.peerLifetime * time.Second)) {
						expiredPeers = append(expiredPeers, peerID)
					}
				}
				if len(expiredPeers) == 0 {
					log.Debug().
						Str("info_hash", hex.EncodeToString(ih[:])).
						Msg("expired peers not found")
					return
				}

				for _, peer := range expiredPeers {
					delete(t.Peers, peer)
				}

				log.Debug().
					Str("info_hash", hex.EncodeToString(ih[:])).
					Msg("expired peers deleted")
			})

		case error:
			log.Error().Err(e)
		}
	}

	return nil
}
