package tracker

import (
	"bufio"
	"bytes"
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

type scrapeResponse struct {
	action        uint32
	transactionID uint32
	seeders       []uint32
	completed     []uint32
	leechers      []uint32
}

func (r *scrapeResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)

	_ = binary.Write(writer, binary.BigEndian, r.seeders)
	_ = binary.Write(writer, binary.BigEndian, r.completed)
	_ = binary.Write(writer, binary.BigEndian, r.leechers)
	writer.Flush()

	return buffer.Bytes()
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

	res := &scrapeResponse{
		action:        2,
		transactionID: req.transactionID,
		seeders:       []uint32{},
		completed:     []uint32{},
		leechers:      []uint32{},
	}

	// todo: cache

	for idx, infoHash := range req.infoHashes {
		if 8+idx*12 >= 768 {
			break
		}

		torrent, ok := torrents.Get(infoHash)
		if !ok {
			res := &errorResponse{
				action:        3,
				transactionID: req.transactionID,
				message:       "Torrent not found",
			}
			pack := res.pack()

			_, err = conn.WriteToUDP(pack, addr)
			if err != nil {
				log.Error().Err(err).Msg("cant write udp scrape error")
				return
			}
			return
		}

		leechers, seeders, completed, _ := torrent.state()
		res.seeders = append(res.seeders, seeders)
		res.completed = append(res.completed, completed)
		res.leechers = append(res.leechers, leechers)
	}
	pack := res.pack()

	_, err = conn.WriteToUDP(pack, addr)
	if err != nil {
		log.Error().Err(err).Msg("cant write udp scrape response")
		return
	}
}
