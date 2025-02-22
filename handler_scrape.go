package tracker

import (
	"encoding/binary"
	"net"

	"github.com/rs/zerolog/log"
)

type scrapeRequest struct {
	connectionID  uint64
	action        uint32
	transactionID uint32
	infoHashes    [][20]byte
}

func (r *scrapeRequest) unpack(bytes []byte) error {
	if len(bytes) < 20 {
		return errorSizeMismatch
	}

	r.connectionID = binary.BigEndian.Uint64(bytes[0:8])
	r.action = binary.BigEndian.Uint32(bytes[8:12])
	r.transactionID = binary.BigEndian.Uint32(bytes[12:16])

	hashes := [][20]byte{}
	for i := 16; i < len(bytes); i = i + 20 {
		infoHash := [20]byte(bytes[i : i+20])
		hashes = append(hashes, infoHash)
	}

	r.infoHashes = hashes

	return nil
}

func handleScrape(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, torrents Storer[InfoHash, *Torrent]) {
	req := &scrapeRequest{}
	err := req.unpack(request)

	if err != nil {
		connectionID := binary.BigEndian.Uint64(request[0:8])
		log.Error().Err(err).
			Int64("connection_id", int64(connectionID)).
			Int("size", len(request)).
			Msg("cant unpack scrape request")
		return
	}

}
