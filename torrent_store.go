package tracker

import (
	"context"
	"time"
)

// TorrentStore persists torrents and their peers. GetTorrent and GetPeer
// return (nil, nil) when the requested row does not exist.
type TorrentStore interface {
	RegisterTorrent(ctx context.Context, t *Torrent) error
	GetTorrent(ctx context.Context, infoHash InfoHash) (*Torrent, error)
	GetPeer(ctx context.Context, infoHash InfoHash, peerID PeerID) (*Peer, error)
	UpsertPeer(ctx context.Context, infoHash InfoHash, peerID PeerID, peer *Peer) error
	IncrementCompleted(ctx context.Context, infoHash InfoHash) error
	ListTorrents(ctx context.Context) (map[InfoHash]*Torrent, error)
	DeleteExpiredPeers(ctx context.Context, before time.Time) (int64, error)
	Close() error
}
