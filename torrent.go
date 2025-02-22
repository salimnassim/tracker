package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

type Peer struct {
	Downloaded uint64
	Left       uint64
	Uploaded   uint64
	Event      uint32
	IP         uint32
	Port       uint16
}

func (r *Peer) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.IP)
	_ = binary.Write(writer, binary.BigEndian, r.Port)
	writer.Flush()

	return buffer.Bytes()
}

type Torrent struct {
	Peers map[[20]byte]*Peer
}

func NewTorrent() *Torrent {
	return &Torrent{
		Peers: map[[20]byte]*Peer{},
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
