package tracker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"log/slog"
	"net"
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

func handleScrape(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, torrents TorrentStore) {
	req := &scrapeRequest{}
	err := req.unpack(request)
	if err != nil {
		connectionID := binary.BigEndian.Uint64(request[0:8])
		slog.Error("cant unpack scrape request",
			"error", err,
			"connection_id", connectionID,
			"size", len(request))
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

		torrent, err := torrents.GetTorrent(context.Background(), InfoHash(infoHash))
		if err != nil {
			slog.Error("cant get torrent for scrape", "error", err)
			return
		}
		if torrent == nil {
			res.seeders = append(res.seeders, 0)
			res.completed = append(res.completed, 0)
			res.leechers = append(res.leechers, 0)
			continue
		}

		leechers, seeders, completed, _ := torrent.state()
		res.seeders = append(res.seeders, seeders)
		res.completed = append(res.completed, completed)
		res.leechers = append(res.leechers, leechers)
	}
	pack := res.pack()

	_, err = conn.WriteToUDP(pack, addr)
	if err != nil {
		slog.Error("cant write udp scrape response", "error", err)
		return
	}
}
