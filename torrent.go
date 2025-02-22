package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

type Peer struct {
	Downloaded uint64 `json:"downloaded"`
	Left       uint64 `json:"left"`
	Uploaded   uint64 `json:"uploaded"`
	Event      uint32 `json:"event"`
	IP         uint32 `json:"ip"`
	Port       uint16 `json:"port"`
}

func (r *Peer) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.IP)
	_ = binary.Write(writer, binary.BigEndian, r.Port)
	writer.Flush()

	return buffer.Bytes()
}

type InfoHash [20]byte

func (h InfoHash) MarshalText() (text []byte, err error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

type PeerID [20]byte

func (h PeerID) MarshalText() (text []byte, err error) {
	return []byte(string(h[:])), nil
}

type Torrent struct {
	Peers map[PeerID]*Peer `json:"peers"`
}

func NewTorrent() *Torrent {
	return &Torrent{
		Peers: map[PeerID]*Peer{},
	}
}

func (r *Torrent) state() (leechers uint32, seeders uint32, peers [][6]byte) {
	for _, peer := range r.Peers {
		if peer.Left > 0 {
			leechers++
		}
		if peer.Left == 0 {
			seeders++
		}
		peers = append(peers, [6]byte(peer.pack()))
	}

	return leechers, seeders, peers
}
