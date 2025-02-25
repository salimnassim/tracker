package tracker

import (
	"context"
	"encoding/hex"
	"os"
	"strconv"
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

type eventCacheLifetime struct{}
type eventPeerLifetime struct{}

type Serverer interface {
	Serve(state chan any, conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent], cache Cacher[InfoHash, *Torrent])
	Address() string
	Port() int
}

type server struct {
	ctx   context.Context
	state chan any

	conns Storer[uint64, time.Time]

	torrents Storer[InfoHash, *Torrent]
	cache    Cacher[InfoHash, *Torrent]
}

func NewServer(conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent], cache Cacher[InfoHash, *Torrent]) *server {
	if conns == nil {
		log.Fatal().
			Msg("server conns store is nil")
	}
	if torrents == nil {
		log.Fatal().
			Msg("server torrents store is nil")
	}
	if cache == nil {
		log.Fatal().
			Msg("server cache is nil")
	}

	return &server{
		ctx:      context.Background(),
		conns:    conns,
		torrents: torrents,
		cache:    cache,
		state:    make(chan any),
	}
}

func (s *server) Start(servers []Serverer) error {
	for _, server := range servers {
		go server.Serve(s.state, s.conns, s.torrents, s.cache)
		log.Info().
			Str("address", server.Address()).
			Int("port", server.Port()).
			Msg("started server")
	}

	// update cache every n seconds
	go func(s *server) {
		cacheInterval, err := strconv.Atoi(os.Getenv("CACHE_INTERVAL"))
		if err != nil {
			log.Fatal().Err(err).Msg("cant parse cache interval")
		}

		cacheTicker := time.NewTicker(time.Duration(cacheInterval) * time.Second)
		for range cacheTicker.C {
			s.state <- eventCacheLifetime{}
		}
	}(s)

	peerLifetime, err := strconv.Atoi(os.Getenv("PEER_LIFETIME"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse peer lifetime")
	}

	// remove expired peers every n seconds
	go func(s *server) {
		peerInterval, err := strconv.Atoi(os.Getenv("PEER_INTERVAL"))
		if err != nil {
			log.Fatal().Err(err).Msg("cant parse peer interval")
		}

		peerTicker := time.NewTicker(time.Duration(peerInterval) * time.Second)
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
			torrent := NewTorrent(e.InfoHash)
			s.torrents.Set(e.InfoHash, torrent)

			log.Info().
				Str("info_hash", hex.EncodeToString(e.InfoHash[:])).
				Msg("torrent registered")

		case eventCacheLifetime:
			s.cache.FromStore(s.torrents)
			log.Info().
				Int("size", s.cache.Size()).
				Msg("cache updated")

		case eventPeerLifetime:
			s.torrents.Map(func(ih InfoHash, t *Torrent) {
				expiredPeers := []PeerID{}
				for peerID, peer := range t.Peers {
					if time.Now().After(peer.Time.Add(time.Duration(peerLifetime) * time.Second)) {
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
