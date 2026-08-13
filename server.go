package tracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
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
	Serve(config *config, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore)
}

type server struct {
	ctx   context.Context
	state chan any

	conns Storer[uint64, time.Time]

	torrents TorrentStore
}

func NewServer(conns Storer[uint64, time.Time], torrents TorrentStore) (*server, error) {
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
		slog.Info("started server", "type", fmt.Sprintf("%T", server))
	}

	go func(s *server) {
		slog.Info("started peer ticker")
		peerTicker := time.NewTicker(config.peerInterval * time.Second)
		for range peerTicker.C {
			s.state <- eventPeerLifetime{}
		}
	}(s)

	for event := range s.state {
		switch e := event.(type) {
		case eventConnection:
			s.conns.Set(e.ConnectionID, time.Now().UTC())

			slog.Info("connection created", "connection_id", e.ConnectionID)
		case eventAnnounce:
			slog.Info("announce", "peer_id", e.PeerId.String())

			torrent, err := s.torrents.GetTorrent(s.ctx, e.InfoHash)
			if err != nil {
				slog.Error("announce get torrent failed",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String(),
					"error", err)
				continue
			}
			if torrent == nil {
				slog.Error("announce torrent not found",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String())
				continue
			}

			existingPeer, err := s.torrents.GetPeer(s.ctx, e.InfoHash, e.PeerId)
			if err != nil {
				slog.Error("announce get peer failed",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String(),
					"error", err)
				continue
			}

			if existingPeer == nil {
				peer := &Peer{
					Event:      e.Event,
					Left:       e.Left,
					Downloaded: e.Downloaded,
					Uploaded:   e.Uploaded,
					IP:         e.IP,
					Port:       e.Port,
					Key:        e.Key,
					Time:       time.Now().UTC(),
				}
				if err := s.torrents.UpsertPeer(s.ctx, e.InfoHash, e.PeerId, peer); err != nil {
					slog.Error("announce create peer failed",
						"info_hash", e.InfoHash.String(),
						"peer_id", e.PeerId.String(),
						"error", err)
					continue
				}

				slog.Info("announce peer created",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String())
				continue
			}

			if existingPeer.Key != 0 && existingPeer.Key != e.Key {
				slog.Info("announce peer key mismatch",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String())
				continue
			}

			if e.Event == 1 {
				if err := s.torrents.IncrementCompleted(s.ctx, e.InfoHash); err != nil {
					slog.Error("announce increment completed failed",
						"info_hash", e.InfoHash.String(),
						"peer_id", e.PeerId.String(),
						"error", err)
					continue
				}

				slog.Info("announce completed",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String())
			}

			updatedPeer := &Peer{
				Event:      e.Event,
				Left:       e.Left,
				Downloaded: e.Downloaded,
				Uploaded:   e.Uploaded,
				IP:         e.IP,
				Port:       e.Port,
				Key:        existingPeer.Key,
				Time:       time.Now().UTC(),
			}
			if err := s.torrents.UpsertPeer(s.ctx, e.InfoHash, e.PeerId, updatedPeer); err != nil {
				slog.Error("announce update peer failed",
					"info_hash", e.InfoHash.String(),
					"peer_id", e.PeerId.String(),
					"error", err)
				continue
			}

			slog.Info("peer updated",
				"info_hash", e.InfoHash.String(),
				"peer_id", e.PeerId.String())

		case eventRegisterTorrent:
			torrent := NewTorrent(e.InfoHash, config)
			if err := s.torrents.RegisterTorrent(s.ctx, torrent); err != nil {
				slog.Error("torrent register failed",
					"info_hash", InfoHash(e.InfoHash).String(),
					"error", err)
				continue
			}

			slog.Info("torrent registered", "info_hash", InfoHash(e.InfoHash).String())

		case eventPeerLifetime:
			cutoff := time.Now().Add(-config.peerLifetime * time.Second)
			deleted, err := s.torrents.DeleteExpiredPeers(s.ctx, cutoff)
			if err != nil {
				slog.Error("expired peer cleanup failed", "error", err)
				continue
			}
			slog.Debug("expired peers deleted", "count", deleted)

		case error:
			slog.Error("server error", "error", e)
		}
	}

	return nil
}
