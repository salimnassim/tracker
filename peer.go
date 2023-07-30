package tracker

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/gofrs/uuid"
)

type Peer struct {
	TorrentID  uuid.UUID `db:"parent_id"`
	PeerID     []byte    `db:"peer_id" validate:"require,len=20"`
	Port       int       `db:"port" validate:"required,number"`
	Uploaded   int       `db:"uploaded" validate:"number"`
	Downloaded int       `db:"downloaded" validate:"number"`
	Left       int       `db:"left" validate:"number"`
	Key        string    `db:"key" validate:"ascii"`
	IP         string    `db:"ip" validate:"ip"`
	Time       time.Time `db:"time"`
	Event      string    `db:"event" validate:"ascii"`
}

func (peer *Peer) Marshal() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.BigEndian, binary.BigEndian.Uint32(
		(net.ParseIP(peer.IP).To4()),
	))

	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, uint16(peer.Port))
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
