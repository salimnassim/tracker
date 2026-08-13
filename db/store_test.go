package db

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/salimnassim/tracker"
)

func newTestStore(t *testing.T) *TorrentStore {
	t.Helper()

	dir := t.TempDir()
	sqlDB, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	applyTestSchema(t, sqlDB)

	return NewTorrentStore(sqlDB)
}

func applyTestSchema(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	content, err := fs.ReadFile(tracker.Migrations, "migrations/00001_init.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	upSQL, _, _ := strings.Cut(string(content), "-- +goose Down")
	if _, err := sqlDB.Exec(upSQL); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

func testInfoHash(b byte) tracker.InfoHash {
	var ih tracker.InfoHash
	for i := range ih {
		ih[i] = b
	}
	return ih
}

func testPeerID(b byte) tracker.PeerID {
	var pid tracker.PeerID
	for i := range pid {
		pid[i] = b
	}
	return pid
}

func TestTorrentStoreRegisterAndGet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	infoHash := testInfoHash(0xAA)
	err := store.RegisterTorrent(ctx, &tracker.Torrent{
		InfoHash: infoHash,
		Magnet:   "magnet:?xt=urn:btih:aa",
	})
	if err != nil {
		t.Fatalf("register torrent: %v", err)
	}

	got, err := store.GetTorrent(ctx, infoHash)
	if err != nil {
		t.Fatalf("get torrent: %v", err)
	}
	if got == nil {
		t.Fatal("expected torrent, got nil")
	}
	if got.Magnet != "magnet:?xt=urn:btih:aa" {
		t.Errorf("magnet = %q", got.Magnet)
	}
	if len(got.Peers) != 0 {
		t.Errorf("expected no peers, got %d", len(got.Peers))
	}

	missing, err := store.GetTorrent(ctx, testInfoHash(0xBB))
	if err != nil {
		t.Fatalf("get missing torrent: %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for missing torrent, got %+v", missing)
	}
}

func TestTorrentStoreUpsertPeer(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	infoHash := testInfoHash(0xCC)
	if err := store.RegisterTorrent(ctx, &tracker.Torrent{InfoHash: infoHash, Magnet: "m"}); err != nil {
		t.Fatalf("register torrent: %v", err)
	}

	peerID := testPeerID(0x01)
	missing, err := store.GetPeer(ctx, infoHash, peerID)
	if err != nil {
		t.Fatalf("get missing peer: %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for missing peer, got %+v", missing)
	}

	peer := &tracker.Peer{Downloaded: 1, Left: 2, Uploaded: 3, Event: 2, Key: 42, Time: time.Now().UTC()}
	if err := store.UpsertPeer(ctx, infoHash, peerID, peer); err != nil {
		t.Fatalf("upsert peer (create): %v", err)
	}

	got, err := store.GetPeer(ctx, infoHash, peerID)
	if err != nil {
		t.Fatalf("get peer: %v", err)
	}
	if got == nil || got.Key != 42 || got.Downloaded != 1 {
		t.Fatalf("unexpected peer state: %+v", got)
	}

	// Upsert again with different stats but same peer/torrent key -- should
	// update in place rather than duplicate.
	peer.Downloaded = 100
	if err := store.UpsertPeer(ctx, infoHash, peerID, peer); err != nil {
		t.Fatalf("upsert peer (update): %v", err)
	}

	torrent, err := store.GetTorrent(ctx, infoHash)
	if err != nil {
		t.Fatalf("get torrent: %v", err)
	}
	if len(torrent.Peers) != 1 {
		t.Fatalf("expected exactly 1 peer after re-upsert, got %d", len(torrent.Peers))
	}
	if torrent.Peers[peerID].Downloaded != 100 {
		t.Errorf("downloaded = %d, want 100", torrent.Peers[peerID].Downloaded)
	}
}

func TestTorrentStoreIncrementCompleted(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	infoHash := testInfoHash(0xDD)
	if err := store.RegisterTorrent(ctx, &tracker.Torrent{InfoHash: infoHash, Magnet: "m"}); err != nil {
		t.Fatalf("register torrent: %v", err)
	}

	if err := store.IncrementCompleted(ctx, infoHash); err != nil {
		t.Fatalf("increment completed: %v", err)
	}
	if err := store.IncrementCompleted(ctx, infoHash); err != nil {
		t.Fatalf("increment completed: %v", err)
	}

	got, err := store.GetTorrent(ctx, infoHash)
	if err != nil {
		t.Fatalf("get torrent: %v", err)
	}
	if got.Completed != 2 {
		t.Errorf("completed = %d, want 2", got.Completed)
	}
}

func TestTorrentStoreDeleteExpiredPeers(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	infoHash := testInfoHash(0xEE)
	if err := store.RegisterTorrent(ctx, &tracker.Torrent{InfoHash: infoHash, Magnet: "m"}); err != nil {
		t.Fatalf("register torrent: %v", err)
	}

	stalePeer := testPeerID(0x01)
	freshPeer := testPeerID(0x02)

	if err := store.UpsertPeer(ctx, infoHash, stalePeer, &tracker.Peer{Time: time.Now().Add(-2 * time.Hour)}); err != nil {
		t.Fatalf("upsert stale peer: %v", err)
	}
	if err := store.UpsertPeer(ctx, infoHash, freshPeer, &tracker.Peer{Time: time.Now()}); err != nil {
		t.Fatalf("upsert fresh peer: %v", err)
	}

	deleted, err := store.DeleteExpiredPeers(ctx, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("delete expired peers: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	torrent, err := store.GetTorrent(ctx, infoHash)
	if err != nil {
		t.Fatalf("get torrent: %v", err)
	}
	if len(torrent.Peers) != 1 {
		t.Fatalf("expected 1 remaining peer, got %d", len(torrent.Peers))
	}
	if _, ok := torrent.Peers[freshPeer]; !ok {
		t.Error("expected fresh peer to remain")
	}
}
