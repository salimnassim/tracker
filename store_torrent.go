package tracker

import (
	"context"

	"github.com/gofrs/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TorrentStorable interface {
	AddTorrent(context.Context, []byte) (*Torrent, error)
	UpsertPeer(context.Context, uuid.UUID, AnnounceRequest) error
	GetTorrent(context.Context, []byte) (*Torrent, error)
	Delete()
	Update()
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

func (ts *TorrentStore) UpsertPeer(ctx context.Context, torrentID uuid.UUID, peer AnnounceRequest) error {
	query := `insert into peers (id, torrent_id, peer_id, ip, port, uploaded, downloaded, "left", event, updated_at)
	values (gen_random_uuid(), $1, $2, $3, $4::integer, $5::integer, $6::integer, $7::integer, $8, now())
	on conflict (torrent_id, peer_id) do update set
	"left" = $9::integer, uploaded = $10::integer, downloaded = $11::integer, updated_at = now(), event = $12`

	_, err := ts.pool.Exec(ctx, query,
		torrentID, peer.PeerID, peer.IP, peer.Port, peer.Uploaded, peer.Downloaded, peer.Left, peer.Event,
		peer.Left, peer.Uploaded, peer.Downloaded, peer.Event)
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

func (ts *TorrentStore) Delete() {

}

func (ts *TorrentStore) Update() {

}
