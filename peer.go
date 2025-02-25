package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"time"
)

type PeerID [20]byte

func (h PeerID) MarshalText() (text []byte, err error) {
	return []byte(string(h[:])), nil
}

type Peer struct {
	Downloaded uint64    `json:"downloaded"`
	Left       uint64    `json:"left"`
	Uploaded   uint64    `json:"uploaded"`
	Event      uint32    `json:"event"`
	IP         uint32    `json:"ip"`
	Port       uint16    `json:"port"`
	Key        uint32    `json:"key"`
	Time       time.Time `json:"time"`
}

func (p *Peer) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, p.IP)
	_ = binary.Write(writer, binary.BigEndian, p.Port)
	writer.Flush()

	return buffer.Bytes()
}

func (p *Peer) MarshalJSON() ([]byte, error) {
	type alias Peer
	type dto struct {
		*alias
		IP   uint32    `json:"ip,omitempty"`
		Port uint16    `json:"port,omitempty"`
		Key  uint32    `json:"key,omitempty"`
		Time time.Time `json:"time,omitzero"`
	}

	return json.Marshal(&dto{
		alias: (*alias)(p),
	})
}
