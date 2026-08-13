package tracker

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeTorrentStore struct {
	mu       sync.Mutex
	torrents map[InfoHash]*Torrent
	peers    map[InfoHash]map[PeerID]*Peer

	deleteExpiredCalls []time.Time
}

func newFakeTorrentStore() *fakeTorrentStore {
	return &fakeTorrentStore{
		torrents: make(map[InfoHash]*Torrent),
		peers:    make(map[InfoHash]map[PeerID]*Peer),
	}
}

func (f *fakeTorrentStore) RegisterTorrent(ctx context.Context, t *Torrent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *t
	f.torrents[t.InfoHash] = &cp
	if _, ok := f.peers[t.InfoHash]; !ok {
		f.peers[t.InfoHash] = make(map[PeerID]*Peer)
	}
	return nil
}

func (f *fakeTorrentStore) GetTorrent(ctx context.Context, infoHash InfoHash) (*Torrent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.torrents[infoHash]
	if !ok {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (f *fakeTorrentStore) GetPeer(ctx context.Context, infoHash InfoHash, peerID PeerID) (*Peer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	peers, ok := f.peers[infoHash]
	if !ok {
		return nil, nil
	}
	p, ok := peers[peerID]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (f *fakeTorrentStore) UpsertPeer(ctx context.Context, infoHash InfoHash, peerID PeerID, peer *Peer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.peers[infoHash]; !ok {
		f.peers[infoHash] = make(map[PeerID]*Peer)
	}
	cp := *peer
	f.peers[infoHash][peerID] = &cp
	return nil
}

func (f *fakeTorrentStore) IncrementCompleted(ctx context.Context, infoHash InfoHash) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.torrents[infoHash]; ok {
		t.Completed++
	}
	return nil
}

func (f *fakeTorrentStore) ListTorrents(ctx context.Context) (map[InfoHash]*Torrent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make(map[InfoHash]*Torrent, len(f.torrents))
	for k, v := range f.torrents {
		cp := *v
		result[k] = &cp
	}
	return result, nil
}

func (f *fakeTorrentStore) DeleteExpiredPeers(ctx context.Context, before time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteExpiredCalls = append(f.deleteExpiredCalls, before)
	var count int64
	for _, peers := range f.peers {
		for id, p := range peers {
			if p.Time.Before(before) {
				delete(peers, id)
				count++
			}
		}
	}
	return count, nil
}

func (f *fakeTorrentStore) Close() error { return nil }

func (f *fakeTorrentStore) getPeer(infoHash InfoHash, peerID PeerID) (*Peer, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	peers, ok := f.peers[infoHash]
	if !ok {
		return nil, false
	}
	p, ok := peers[peerID]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

func (f *fakeTorrentStore) getTorrent(infoHash InfoHash) (*Torrent, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.torrents[infoHash]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func startTestServer(t *testing.T, fake *fakeTorrentStore, peerLifetime time.Duration) (*server, *config) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	conns := NewStore[uint64, time.Time]()
	srv, err := NewServer(ctx, conns, fake)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	cfg := NewConfig("", 0, "", 0, "udp://localhost:6881", 3600, peerLifetime, "")

	go func() {
		_ = srv.Start(cfg, []Serverer{})
	}()

	return srv, cfg
}

func TestEventLoopAnnounceCreatesPeer(t *testing.T) {
	fake := newFakeTorrentStore()
	srv, _ := startTestServer(t, fake, 3600)

	infoHash := testInfoHash(0x01)
	peerID := testPeerID(0x01)

	srv.state <- eventRegisterTorrent{InfoHash: infoHash}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getTorrent(infoHash)
		return ok
	})

	srv.state <- eventAnnounce{
		InfoHash:   infoHash,
		PeerId:     peerID,
		Downloaded: 10,
		Left:       100,
		Event:      2,
		Key:        42,
	}

	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getPeer(infoHash, peerID)
		return ok
	})

	peer, _ := fake.getPeer(infoHash, peerID)
	if peer.Key != 42 || peer.Downloaded != 10 || peer.Left != 100 {
		t.Errorf("unexpected peer state: %+v", peer)
	}
}

func TestEventLoopAnnounceUpdatePreservesKey(t *testing.T) {
	fake := newFakeTorrentStore()
	srv, _ := startTestServer(t, fake, 3600)

	infoHash := testInfoHash(0x02)
	peerID := testPeerID(0x02)

	srv.state <- eventRegisterTorrent{InfoHash: infoHash}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getTorrent(infoHash)
		return ok
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Key: 0, Downloaded: 1}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getPeer(infoHash, peerID)
		return ok
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Key: 999, Downloaded: 999}
	waitForCondition(t, time.Second, func() bool {
		p, ok := fake.getPeer(infoHash, peerID)
		return ok && p.Downloaded == 999
	})

	peer, _ := fake.getPeer(infoHash, peerID)
	if peer.Key != 0 {
		t.Errorf("key = %d, want 0 (preserved from creation, never overwritten)", peer.Key)
	}
}

func TestEventLoopAnnounceKeyMismatchSkipsUpdate(t *testing.T) {
	fake := newFakeTorrentStore()
	srv, _ := startTestServer(t, fake, 3600)

	infoHash := testInfoHash(0x03)
	peerID := testPeerID(0x03)
	otherPeerID := testPeerID(0x04)

	srv.state <- eventRegisterTorrent{InfoHash: infoHash}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getTorrent(infoHash)
		return ok
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Key: 42, Downloaded: 1}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getPeer(infoHash, peerID)
		return ok
	})

	// Mismatched key, should be silently skipped.
	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Key: 999, Downloaded: 555}
	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: otherPeerID, Key: 1}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getPeer(infoHash, otherPeerID)
		return ok
	})

	peer, _ := fake.getPeer(infoHash, peerID)
	if peer.Downloaded != 1 {
		t.Errorf("downloaded = %d, want 1 (mismatched-key update should have been skipped)", peer.Downloaded)
	}
}

func TestEventLoopAnnounceCompletedIncrements(t *testing.T) {
	fake := newFakeTorrentStore()
	srv, _ := startTestServer(t, fake, 3600)

	infoHash := testInfoHash(0x05)
	peerID := testPeerID(0x05)

	srv.state <- eventRegisterTorrent{InfoHash: infoHash}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getTorrent(infoHash)
		return ok
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Event: 2}
	waitForCondition(t, time.Second, func() bool {
		_, ok := fake.getPeer(infoHash, peerID)
		return ok
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Event: 1}
	waitForCondition(t, time.Second, func() bool {
		torrent, ok := fake.getTorrent(infoHash)
		return ok && torrent.Completed == 1
	})

	srv.state <- eventAnnounce{InfoHash: infoHash, PeerId: peerID, Event: 1}
	waitForCondition(t, time.Second, func() bool {
		torrent, ok := fake.getTorrent(infoHash)
		return ok && torrent.Completed == 2
	})
}

func TestEventLoopPeerLifetimeDeletesExpiredPeers(t *testing.T) {
	fake := newFakeTorrentStore()
	srv, cfg := startTestServer(t, fake, 60)

	infoHash := testInfoHash(0x06)
	stalePeer := testPeerID(0x06)
	freshPeer := testPeerID(0x07)

	if err := fake.RegisterTorrent(context.Background(), &Torrent{InfoHash: infoHash}); err != nil {
		t.Fatalf("register torrent: %v", err)
	}
	if err := fake.UpsertPeer(context.Background(), infoHash, stalePeer, &Peer{Time: time.Now().Add(-2 * time.Hour)}); err != nil {
		t.Fatalf("seed stale peer: %v", err)
	}
	if err := fake.UpsertPeer(context.Background(), infoHash, freshPeer, &Peer{Time: time.Now()}); err != nil {
		t.Fatalf("seed fresh peer: %v", err)
	}

	srv.state <- eventPeerLifetime{}

	waitForCondition(t, time.Second, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.deleteExpiredCalls) == 1
	})

	if _, ok := fake.getPeer(infoHash, stalePeer); ok {
		t.Error("expected stale peer to be deleted")
	}
	if _, ok := fake.getPeer(infoHash, freshPeer); !ok {
		t.Error("expected fresh peer to remain")
	}

	fake.mu.Lock()
	cutoff := fake.deleteExpiredCalls[0]
	fake.mu.Unlock()

	wantCutoff := time.Now().Add(-cfg.peerLifetime * time.Second)
	if diff := wantCutoff.Sub(cutoff); diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("cutoff = %v, want close to %v (diff %v)", cutoff, wantCutoff, diff)
	}
}

func testInfoHash(b byte) InfoHash {
	var ih InfoHash
	for i := range ih {
		ih[i] = b
	}
	return ih
}

func testPeerID(b byte) PeerID {
	var pid PeerID
	for i := range pid {
		pid[i] = b
	}
	return pid
}
