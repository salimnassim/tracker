package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

type Peer struct {
	Downloaded uint64 `json:"downloaded"`
	Left       uint64 `json:"left"`
	Uploaded   uint64 `json:"uploaded"`
	Event      uint32 `json:"event"`
	IP         uint32 `json:"ip"`
	Port       uint16 `json:"port"`
	Time       int64  `json:"time"`
}

func (p *Peer) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, p.IP)
	_ = binary.Write(writer, binary.BigEndian, p.Port)
	writer.Flush()

	return buffer.Bytes()
}

// func (p *Peer) MarshalJSON() ([]byte, error) {
// 	type alias Peer
// 	type dto struct {
// 		*alias
// 		IP string `json:"ip"`
// 	}

// 	return json.Marshal(&dto{
// 		alias: (*alias)(p),
// 		IP: fmt.Sprintf("%d.%d.%d.%d",
// 			byte(p.IP>>24),
// 			byte(p.IP>>16),
// 			byte(p.IP>>8),
// 			byte(p.IP),
// 		),
// 	})
// }

type InfoHash [20]byte

func (h InfoHash) MarshalText() (text []byte, err error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

type PeerID [20]byte

func (h PeerID) MarshalText() (text []byte, err error) {
	return []byte(string(h[:])), nil
}

type Torrent struct {
	InfoHash  InfoHash         `json:"-"`
	Peers     map[PeerID]*Peer `json:"peers"`
	Completed uint32           `json:"completed"`
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
