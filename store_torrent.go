package tracker

import (
	"context"

	"github.com/gofrs/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TorrentStorable interface {
	GetTorrent(ctx context.Context, infoHash []byte) (Torrent, error)
	GetTorrents(ctx context.Context) ([]Torrent, error)
	AddTorrent(ctx context.Context, infoHash []byte) (Torrent, error)

	GetPeers(ctx context.Context, torrentID uuid.UUID) ([]Peer, error)
	UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (error, bool)
	InsertOrUpdatePeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error

	Ping(ctx context.Context) (bool, error)
}

type TorrentStore struct {
	pool *pgxpool.Pool
}

func NewTorrentStore(pool *pgxpool.Pool) *TorrentStore {
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

func (ts *TorrentStore) GetPeers(ctx context.Context, torrentID uuid.UUID) ([]Peer, error) {
	query := `select id, torrent_id, peer_id, ip, port, uploaded, downloaded, "left", event, key, updated_at
	from peers
	where torrent_id = $1
	limit 24`

	var peers []Peer
	rows, err := ts.pool.Query(ctx, query, torrentID)
	if err != nil {
		return []Peer{}, err
	}

	peers, err = pgx.CollectRows[Peer](rows, pgx.RowToStructByName[Peer])
	if err != nil {
		return []Peer{}, err
	}
	return peers, nil
}

func (ts *TorrentStore) UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (error, bool) {
	query := `update peers set
	peer_id = $1, ip = $2, port = $3, uploaded = $4, downloaded = $5, "left" = $6, event = $7, updated_at = now()
	where torrent_id = $8 and key = $9`

	tag, err := ts.pool.Exec(ctx, query,
		req.PeerID, req.IP, req.Port, req.Uploaded, req.Downloaded, req.Left, req.Event,
		torrentID, req.Key)
	if err != nil {
		return err, false
	}

	return nil, (tag.RowsAffected() > 0)
}

func (ts *TorrentStore) InsertOrUpdatePeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error {
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

func (ts *TorrentStore) GetTorrents(ctx context.Context) ([]Torrent, error) {
	query := `select t.id, t.info_hash, t.completed, t.created_at,
		(select count(*) from peers where peers.torrent_id = t.id and peers.left = 0) as seeders,
		(select count(*) from peers where peers.torrent_id = t.id and peers.left != 0) as leechers
	from torrents t`

	var torrents []Torrent
	rows, err := ts.pool.Query(ctx, query)
	if err != nil {
		return []Torrent{}, err
	}

	torrents, err = pgx.CollectRows[Torrent](rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return []Torrent{}, err
	}

	return torrents, nil
}

func (ts *TorrentStore) Ping(ctx context.Context) (bool, error) {
	err := ts.pool.Ping(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
