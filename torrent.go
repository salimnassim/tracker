package tracker

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type InfoHash [20]byte

func (h InfoHash) MarshalText() (text []byte, err error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

type Torrent struct {
	InfoHash  InfoHash         `json:"-"`
	Magnet    string           `json:"magnet"`
	Completed uint32           `json:"completed"`
	Peers     map[PeerID]*Peer `json:"peers"`
}

func (t *Torrent) MarshalJSON() ([]byte, error) {
	type alias Torrent
	type dto struct {
		*alias
	}

	return json.Marshal(&dto{
		alias: (*alias)(t),
	})
}

func NewTorrent(infoHash InfoHash, config *config) *Torrent {
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%x&tr=%s", infoHash, config.udpURL)

	return &Torrent{
		InfoHash:  infoHash,
		Magnet:    magnet,
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
