package tracker

import (
	"context"

	"github.com/gofrs/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TorrentStorable interface {
	AddTorrent(ctx context.Context, infoHash []byte) (*Torrent, error)
	UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (error, bool)
	UpsertPeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error
	GetTorrent(ctx context.Context, infoHash []byte) (*Torrent, error)
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

func (ts *TorrentStore) AddTorrent(ctx context.Context, infoHash []byte) (*Torrent, error) {
	query := `insert into torrents (id, info_hash, completed, created_at)
	values (gen_random_uuid(), $1, 0, now())
	returning id, info_hash, completed, created_at`

	rows, err := ts.pool.Query(ctx, query, infoHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	torrent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return nil, err
	}

	return &torrent, nil
}

func (ts *TorrentStore) UpdatePeerWithKey(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) (error, bool) {
	query := `update peers set
	peer_id = $1, ip = $2, port = $3, uploaded = $4::integer, downloaded = $5::integer, "left" = $6::integer, event = $7, updated_at = now()
	where torrent_id = $8 and key = $9`

	tag, err := ts.pool.Exec(ctx, query,
		req.PeerID, req.IP, req.Port, req.Uploaded, req.Downloaded, req.Left, req.Event,
		torrentID, req.Key)
	if err != nil {
		return err, false
	}

	return nil, (tag.RowsAffected() > 0)
}

func (ts *TorrentStore) UpsertPeer(ctx context.Context, torrentID uuid.UUID, req AnnounceRequest) error {
	query := `insert into peers (id, torrent_id, peer_id, ip, port, uploaded, downloaded, "left", event, key, updated_at)
	values (gen_random_uuid(), $1, $2, $3, $4::integer, $5::integer, $6::integer, $7::integer, $8, $9, now())
	on conflict (torrent_id, peer_id) do update set
	"left" = $10::integer, uploaded = $11::integer, downloaded = $12::integer, updated_at = now(), event = $13`

	_, err := ts.pool.Exec(ctx, query,
		torrentID, req.PeerID, req.IP, req.Port, req.Uploaded, req.Downloaded, req.Left, req.Event, req.Key,
		req.Left, req.Uploaded, req.Downloaded, req.Event)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TorrentStore) GetTorrent(ctx context.Context, infoHash []byte) (*Torrent, error) {
	query := `select id, info_hash, completed, created_at from torrents
	where info_hash = $1
	limit 1`

	var torrent Torrent
	rows, err := ts.pool.Query(ctx, query, infoHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	torrent, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Torrent])
	if err != nil {
		return nil, err
	}

	return &torrent, nil
}

func (ts *TorrentStore) Ping(ctx context.Context) (bool, error) {
	err := ts.pool.Ping(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
