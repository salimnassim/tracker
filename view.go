package tracker

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
)

var templateFuncs = template.FuncMap{
	"safeURL": func(s string) template.URL { return template.URL(s) },
}

var eventLabels = map[uint32]string{
	0: "update",
	1: "completed",
	2: "started",
	3: "stopped",
}

func eventLabel(event uint32) string {
	if label, ok := eventLabels[event]; ok {
		return label
	}
	return strconv.FormatUint(uint64(event), 10)
}

func truncateHash(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func formatBytes(n uint64) string {
	if n == 0 {
		return "0 B"
	}

	units := [...]string{"B", "KB", "MB", "GB", "TB"}
	v := float64(n)
	i := 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}

	decimals := 0
	if v < 10 && i != 0 {
		decimals = 1
	}
	return fmt.Sprintf("%.*f %s", decimals, v, units[i])
}

func peerIDDisplay(id PeerID) string {
	b := make([]byte, len(id))
	for i, c := range id {
		if c >= 0x20 && c <= 0x7e {
			b[i] = c
		} else {
			b[i] = '.'
		}
	}
	return string(b)
}

func countPeers(peers map[PeerID]*Peer) (seeders, leechers int) {
	for _, p := range peers {
		if p.Left == 0 {
			seeders++
		} else {
			leechers++
		}
	}
	return seeders, leechers
}

type torrentListItem struct {
	Hash      string
	HashShort string
	Magnet    string
	Seeders   int
	Leechers  int
	PeerCount int
	Completed uint32
}

type listPageData struct {
	Torrents []torrentListItem
	Summary  string
}

func newListPageData(torrents map[InfoHash]*Torrent) listPageData {
	items := make([]torrentListItem, 0, len(torrents))
	totalPeers := 0
	for hash, t := range torrents {
		seeders, leechers := countPeers(t.Peers)
		peerCount := seeders + leechers
		totalPeers += peerCount

		hashString := hash.String()
		items = append(items, torrentListItem{
			Hash:      hashString,
			HashShort: truncateHash(hashString, 16),
			Magnet:    t.Magnet,
			Seeders:   seeders,
			Leechers:  leechers,
			PeerCount: peerCount,
			Completed: t.Completed,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Hash < items[j].Hash })

	return listPageData{
		Torrents: items,
		Summary:  fmt.Sprintf("%d torrents indexed · %d peers online", len(items), totalPeers),
	}
}

type statPill struct {
	Class string
	Label string
}

type peerView struct {
	ID            string
	IDHex         string
	DownloadedFmt string
	UploadedFmt   string
	LeftFmt       string
	IsSeeding     bool
	StatusLabel   string
	EventLabel    string
}

type detailPageData struct {
	Hash      string
	Magnet    string
	Seeders   int
	Leechers  int
	PeerCount int
	Completed uint32
	Pills     []statPill
	Peers     []peerView
	HasPeers  bool
	ShowEvent bool
}

func newDetailPageData(hash InfoHash, t *Torrent, showEvent bool) detailPageData {
	peerViews := make([]peerView, 0, len(t.Peers))
	seeders, leechers := 0, 0
	for id, p := range t.Peers {
		isSeeding := p.Left == 0
		if isSeeding {
			seeders++
		} else {
			leechers++
		}

		statusLabel := "Leeching"
		if isSeeding {
			statusLabel = "Seeding"
		}

		peerViews = append(peerViews, peerView{
			ID:            peerIDDisplay(id),
			IDHex:         id.String(),
			DownloadedFmt: formatBytes(p.Downloaded),
			UploadedFmt:   formatBytes(p.Uploaded),
			LeftFmt:       formatBytes(p.Left),
			IsSeeding:     isSeeding,
			StatusLabel:   statusLabel,
			EventLabel:    eventLabel(p.Event),
		})
	}
	sort.Slice(peerViews, func(i, j int) bool { return peerViews[i].IDHex < peerViews[j].IDHex })

	peerCount := len(peerViews)

	return detailPageData{
		Hash:      hash.String(),
		Magnet:    t.Magnet,
		Seeders:   seeders,
		Leechers:  leechers,
		PeerCount: peerCount,
		Completed: t.Completed,
		Pills: []statPill{
			{Class: "seeding", Label: fmt.Sprintf("%d seeding", seeders)},
			{Class: "leeching", Label: fmt.Sprintf("%d leeching", leechers)},
			{Class: "peers", Label: fmt.Sprintf("%d peers total", peerCount)},
			{Class: "completed", Label: fmt.Sprintf("%d completed", t.Completed)},
		},
		Peers:     peerViews,
		HasPeers:  peerCount > 0,
		ShowEvent: showEvent,
	}
}
