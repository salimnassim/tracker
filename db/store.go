package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/salimnassim/tracker"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database at path and configures it for a single
// writer / many concurrent readers workload (WAL journal mode).
func Open(path string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("set journal_mode: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)

	return sqlDB, nil
}

// TorrentStore implements tracker.TorrentStore against a SQLite database
// via sqlc-generated queries.
type TorrentStore struct {
	db *sql.DB
	q  *Queries
}

func NewTorrentStore(sqlDB *sql.DB) *TorrentStore {
	return &TorrentStore{
		db: sqlDB,
		q:  New(sqlDB),
	}
}

func (s *TorrentStore) Close() error {
	return s.db.Close()
}

func (s *TorrentStore) RegisterTorrent(ctx context.Context, t *tracker.Torrent) error {
	_, err := s.q.RegisterTorrent(ctx, RegisterTorrentParams{
		InfoHash:  t.InfoHash.String(),
		Magnet:    t.Magnet,
		Completed: int64(t.Completed),
	})
	if err != nil {
		return fmt.Errorf("register torrent: %w", err)
	}
	return nil
}

func (s *TorrentStore) GetTorrent(ctx context.Context, infoHash tracker.InfoHash) (*tracker.Torrent, error) {
	row, err := s.q.GetTorrent(ctx, infoHash.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get torrent: %w", err)
	}

	peerRows, err := s.q.ListPeersByTorrent(ctx, row.InfoHash)
	if err != nil {
		return nil, fmt.Errorf("list peers for torrent: %w", err)
	}

	t := &tracker.Torrent{
		InfoHash:  infoHash,
		Magnet:    row.Magnet,
		Completed: uint32(row.Completed),
		Peers:     make(map[tracker.PeerID]*tracker.Peer, len(peerRows)),
	}
	for _, p := range peerRows {
		peerID, err := decodePeerID(p.PeerID)
		if err != nil {
			return nil, err
		}
		t.Peers[peerID] = peerFromRow(p)
	}
	return t, nil
}

func (s *TorrentStore) GetPeer(ctx context.Context, infoHash tracker.InfoHash, peerID tracker.PeerID) (*tracker.Peer, error) {
	row, err := s.q.GetPeer(ctx, GetPeerParams{
		InfoHash: infoHash.String(),
		PeerID:   peerID.String(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get peer: %w", err)
	}
	return peerFromRow(row), nil
}

func (s *TorrentStore) UpsertPeer(ctx context.Context, infoHash tracker.InfoHash, peerID tracker.PeerID, peer *tracker.Peer) error {
	err := s.q.UpsertPeer(ctx, UpsertPeerParams{
		InfoHash:   infoHash.String(),
		PeerID:     peerID.String(),
		Downloaded: int64(peer.Downloaded),
		Left:       int64(peer.Left),
		Uploaded:   int64(peer.Uploaded),
		Event:      int64(peer.Event),
		Ip:         int64(peer.IP),
		Port:       int64(peer.Port),
		Key:        int64(peer.Key),
		UpdatedAt:  peer.Time,
	})
	if err != nil {
		return fmt.Errorf("upsert peer: %w", err)
	}
	return nil
}

func (s *TorrentStore) IncrementCompleted(ctx context.Context, infoHash tracker.InfoHash) error {
	if err := s.q.IncrementCompleted(ctx, infoHash.String()); err != nil {
		return fmt.Errorf("increment completed: %w", err)
	}
	return nil
}

func (s *TorrentStore) ListTorrents(ctx context.Context) (map[tracker.InfoHash]*tracker.Torrent, error) {
	torrentRows, err := s.q.ListTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list torrents: %w", err)
	}
	peerRows, err := s.q.ListAllPeers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all peers: %w", err)
	}

	result := make(map[tracker.InfoHash]*tracker.Torrent, len(torrentRows))
	for _, row := range torrentRows {
		infoHash, err := decodeInfoHash(row.InfoHash)
		if err != nil {
			return nil, err
		}
		result[infoHash] = &tracker.Torrent{
			InfoHash:  infoHash,
			Magnet:    row.Magnet,
			Completed: uint32(row.Completed),
			Peers:     make(map[tracker.PeerID]*tracker.Peer),
		}
	}

	for _, p := range peerRows {
		infoHash, err := decodeInfoHash(p.InfoHash)
		if err != nil {
			return nil, err
		}
		t, ok := result[infoHash]
		if !ok {
			continue
		}
		peerID, err := decodePeerID(p.PeerID)
		if err != nil {
			return nil, err
		}
		t.Peers[peerID] = peerFromRow(p)
	}

	return result, nil
}

func (s *TorrentStore) DeleteExpiredPeers(ctx context.Context, before time.Time) (int64, error) {
	n, err := s.q.DeleteExpiredPeers(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired peers: %w", err)
	}
	return n, nil
}

func peerFromRow(row Peer) *tracker.Peer {
	return &tracker.Peer{
		Downloaded: uint64(row.Downloaded),
		Left:       uint64(row.Left),
		Uploaded:   uint64(row.Uploaded),
		Event:      uint32(row.Event),
		IP:         uint32(row.Ip),
		Port:       uint16(row.Port),
		Key:        uint32(row.Key),
		Time:       row.UpdatedAt,
	}
}

func decodeInfoHash(s string) (tracker.InfoHash, error) {
	var ih tracker.InfoHash
	b, err := hex.DecodeString(s)
	if err != nil {
		return ih, fmt.Errorf("decode info_hash: %w", err)
	}
	if len(b) != len(ih) {
		return ih, fmt.Errorf("decode info_hash: unexpected length %d", len(b))
	}
	copy(ih[:], b)
	return ih, nil
}

func decodePeerID(s string) (tracker.PeerID, error) {
	var pid tracker.PeerID
	b, err := hex.DecodeString(s)
	if err != nil {
		return pid, fmt.Errorf("decode peer_id: %w", err)
	}
	if len(b) != len(pid) {
		return pid, fmt.Errorf("decode peer_id: unexpected length %d", len(b))
	}
	copy(pid[:], b)
	return pid, nil
}
