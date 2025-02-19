package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

type Peer struct {
	IP   uint32
	Port uint16
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
	Peers []Peer
}

func (r *Torrent) zip() [][8]byte {
	var peers [][8]byte
	for _, peer := range r.Peers {
		peers = append(peers, [8]byte(peer.pack()))
	}

	return peers
}
