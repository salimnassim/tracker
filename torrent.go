package tracker

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

type InfoHash [20]byte

func (h InfoHash) MarshalText() (text []byte, err error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

type Torrent struct {
	InfoHash  InfoHash         `json:"-"`
	Completed uint32           `json:"completed"`
	Peers     map[PeerID]*Peer `json:"peers"`
}

func (t *Torrent) MarshalJSON() ([]byte, error) {
	type alias Torrent
	type dto struct {
		Magnet string `json:"magnet"`
		*alias
	}

	udp_tracker_url := url.QueryEscape(os.Getenv("BT_UDP_URL"))
	return json.Marshal(&dto{
		Magnet: fmt.Sprintf("magnet:?xt=urn:btih:%x&tr=%s", t.InfoHash, udp_tracker_url),
		alias:  (*alias)(t),
	})
}

func NewTorrent(infoHash InfoHash) *Torrent {
	return &Torrent{
		InfoHash:  infoHash,
		Peers:     map[PeerID]*Peer{},
		Completed: 0,
	}
}

func (r *Torrent) state() (leechers uint32, seeders uint32, completed uint32, peers [][6]byte) {
	for _, peer := range r.Peers {
		// skip stopped peers
		if peer.Event == 3 {
			continue
		}
		if peer.Left > 0 {
			leechers++
		}
		if peer.Left == 0 {
			seeders++
		}
		peers = append(peers, [6]byte(peer.pack()))
	}

	return leechers, seeders, r.Completed, peers
}
