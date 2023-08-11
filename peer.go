package tracker

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Peer struct {
	ID         uuid.UUID `db:"id"`
	TorrentID  uuid.UUID `db:"torrent_id"`
	PeerID     []byte    `db:"peer_id"`
	Port       int       `db:"port"`
	Uploaded   int       `db:"uploaded"`
	Downloaded int       `db:"downloaded"`
	Left       int       `db:"left"`
	Key        string    `db:"key"`
	IP         net.IP    `db:"ip"`
	UpdatedAt  time.Time `db:"updated_at"`
	Event      string    `db:"event"`
}

func (peer *Peer) Marshal() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.BigEndian, binary.BigEndian.Uint32(
		(net.ParseIP(peer.IP.String()).To4()),
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

// Tries to return client type based on peer_id
func (peer *Peer) Client() string {

	client := strings.ToLower(string(peer.PeerID))
	if strings.HasPrefix(client, "-lt") {
		return "libtorrent"
	}
	if strings.HasPrefix(client, "-tr") {
		return "Transmission"
	}
	if strings.HasPrefix(client, "-ut") {
		return "µtorrent"
	}
	return "unknown"
}
