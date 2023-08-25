package tracker

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TorrentStorable interface {
	// Add torrent to store.
	AddTorrent(ctx context.Context, infoHash []byte) (Torrent, error)
	// Get torrent from store.
	GetTorrent(ctx context.Context, infoHash []byte) (Torrent, error)
	// Increments torrentID completed property by one.
	IncrementTorrent(ctx context.Context, torrentID uuid.UUID) error
	// Get all torrents in store.
	AllTorrents(ctx context.Context) ([]Torrent, error)
	// Get all peers for torrentID.
	AllPeers(ctx context.Context, torrentID uuid.UUID) ([]Peer, error)
	// Try to update peer which already exist in the store.
	// Operation success is denoted by bool.
	UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (bool, error)
	// Update or insert peer to store.
	UpsertPeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error
	// Remove stale peers that have not announced in interval.
	CleanPeers(ctx context.Context, interval time.Duration) (int, error)
	// Log announce request.
	Log(ctx context.Context, req AnnounceRequest) error
	// Test store connection.
	Ping(ctx context.Context) (bool, error)
}

type TorrentStore struct {
	pool *pgxpool.Pool
}

func NewTorrentStore(pool *pgxpool.Pool) *TorrentStore {
	// todo: create pgxpool here
	return &TorrentStore{
		pool: pool,
	}
}

func (ts *TorrentStore) AddTorrent(ctx context.Context, infoHash []byte) (Torrent, error) {
	query := `insert into torrents (id, info_hash, completed, created_at)
	values (gen_random_uuid(), $1, 0, now())
	returning id, info_hash, completed, created_at, 0 as seeders, 0 as leechers`

	rows, err := ts.pool.Query(ctx, query, infoHash)
	if err != nil {
		return Torrent{}, err
	}
	defer rows.Close()

	torrent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return Torrent{}, err
	}

	return torrent, nil
}

func (ts *TorrentStore) IncrementTorrent(ctx context.Context, torrentID uuid.UUID) error {
	query := `update torrents
	set completed = completed + 1
	where id = $1`

	_, err := ts.pool.Exec(ctx, query, torrentID)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TorrentStore) AllPeers(ctx context.Context, torrentID uuid.UUID) ([]Peer, error) {
	query := `select id, torrent_id, peer_id, ip, port, uploaded, downloaded, "left", event, key, updated_at
	from peers
	where torrent_id = $1
	limit 24`

	var peers []Peer
	rows, err := ts.pool.Query(ctx, query, torrentID)
	if err != nil {
		return []Peer{}, err
	}
	defer rows.Close()

	peers, err = pgx.CollectRows[Peer](rows, pgx.RowToStructByName[Peer])
	if err != nil {
		return []Peer{}, err
	}
	return peers, nil
}

func (ts *TorrentStore) UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (bool, error) {
	query := `update peers set
	peer_id = $1, ip = $2, port = $3, uploaded = $4, downloaded = $5, "left" = $6, event = $7, updated_at = now()
	where torrent_id = $8 and key = $9`

	tag, err := ts.pool.Exec(ctx, query,
		req.PeerID, req.IP, req.Port, req.Uploaded, req.Downloaded, req.Left, req.Event,
		torrentID, req.Key)
	if err != nil {
		return false, err
	}

	return (tag.RowsAffected() > 0), nil
}

func (ts *TorrentStore) UpsertPeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error {
	query := `insert into peers (id, torrent_id, peer_id, ip, port, uploaded, downloaded, "left", event, key, updated_at)
	values (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, now())
	on conflict (torrent_id, peer_id) do update set
	"left" = $10, uploaded = $11, downloaded = $12, updated_at = now(), event = $13`

	_, err := ts.pool.Exec(ctx, query,
		torrentID, req.PeerID, req.IP, req.Port, req.Uploaded, req.Downloaded, req.Left, req.Event, req.Key,
		req.Left, req.Uploaded, req.Downloaded, req.Event)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TorrentStore) GetTorrent(ctx context.Context, infoHash []byte) (Torrent, error) {
	query := `select t.id, t.info_hash, t.completed, t.created_at, 
		(select count(*) from peers where peers.torrent_id = t.id and peers.left = 0) as seeders,
		(select count(*) from peers where peers.torrent_id = t.id and peers.left != 0) as leechers
	from torrents t
	where t.info_hash = $1
	limit 1`

	var torrent Torrent
	rows, err := ts.pool.Query(ctx, query, infoHash)
	if err != nil {
		return Torrent{}, err
	}
	defer rows.Close()

	torrent, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return Torrent{}, err
	}

	return torrent, nil
}

func (ts *TorrentStore) AllTorrents(ctx context.Context) ([]Torrent, error) {
	query := `select t.id, t.info_hash, t.completed, t.created_at,
		(select count(*) from peers where peers.torrent_id = t.id and peers.left = 0) as seeders,
		(select count(*) from peers where peers.torrent_id = t.id and peers.left != 0) as leechers
	from torrents t`

	var torrents []Torrent
	rows, err := ts.pool.Query(ctx, query)
	if err != nil {
		return []Torrent{}, err
	}
	defer rows.Close()

	torrents, err = pgx.CollectRows[Torrent](rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return []Torrent{}, err
	}

	return torrents, nil
}

// Removers stale peers that have not updated in a x duration.
func (ts *TorrentStore) CleanPeers(ctx context.Context, interval time.Duration) (int, error) {
	query := `delete from peers	where updated_at < now() - $1::interval`

	tag, err := ts.pool.Exec(ctx, query, interval)
	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}

func (ts *TorrentStore) Log(ctx context.Context, req AnnounceRequest) error {
	query := `insert into announce_log (id, info_hash, peer_id, event, ip, port, key, uploaded, downloaded, "left", created_at)
	values (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, now())`

	_, err := ts.pool.Exec(ctx, query, req.InfoHash, req.PeerID, req.Event, req.IP, req.Port, req.Key, req.Uploaded, req.Downloaded, req.Left)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TorrentStore) Ping(ctx context.Context) (bool, error) {
	err := ts.pool.Ping(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
