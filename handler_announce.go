package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	"github.com/rs/zerolog/log"
)

type announceRequest struct {
	connectionID  uint64
	action        uint32
	transactionID uint32
	infoHash      [20]byte
	peerID        [20]byte
	downloaded    uint64
	left          uint64
	uploaded      uint64
	event         uint32
	ip            uint32
	key           uint32
	numWant       uint32
	port          uint16
}

func (r *announceRequest) unpack(bytes []byte) error {
	if len(bytes) < 98 || len(bytes) > 120 {
		return errorSizeMismatch
	}

	r.connectionID = binary.BigEndian.Uint64(bytes[0:8])
	r.action = binary.BigEndian.Uint32(bytes[8:12])
	r.transactionID = binary.BigEndian.Uint32(bytes[12:16])
	r.infoHash = [20]byte(bytes[16:36])
	r.peerID = [20]byte(bytes[36:56])
	r.downloaded = binary.BigEndian.Uint64(bytes[56:64])
	r.left = binary.BigEndian.Uint64(bytes[64:72])
	r.uploaded = binary.BigEndian.Uint64(bytes[72:80])
	r.event = binary.BigEndian.Uint32(bytes[80:84])
	r.ip = binary.BigEndian.Uint32(bytes[84:88])
	r.key = binary.BigEndian.Uint32(bytes[88:92])
	r.numWant = binary.BigEndian.Uint32(bytes[92:96])
	r.port = binary.BigEndian.Uint16(bytes[96:98])

	// todo: extended bytes

	return nil
}

type announceResponse struct {
	action        uint32
	transactionID uint32
	interval      uint32
	leechers      uint32
	seeders       uint32
	peers         [][6]byte
}

func (r *announceResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, r.interval)
	_ = binary.Write(writer, binary.BigEndian, r.leechers)
	_ = binary.Write(writer, binary.BigEndian, r.seeders)
	_ = binary.Write(writer, binary.BigEndian, r.peers)
	writer.Flush()

	return buffer.Bytes()
}

func handleAnnounce(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, torrents Storer[InfoHash, *Torrent]) {
	req := &announceRequest{}
	err := req.unpack(request)
	if err != nil {
		connectionID := binary.BigEndian.Uint64(request[0:8])
		log.Error().Err(err).
			Int64("connection_id", int64(connectionID)).
			Int("size", len(request)).
			Msg("cant unpack announce request")
		return
	}

	ip := addr.IP.To4()
	if ip == nil {
		log.Error().Msg("invalid ipv4 address")
		return
	}

	torrent, ok := torrents.Get(req.infoHash)
	if !ok {
		state <- eventRegisterTorrent{
			InfoHash: req.infoHash,
		}

		state <- eventAnnounce{
			InfoHash:   req.infoHash,
			PeerId:     req.peerID,
			Downloaded: req.downloaded,
			Left:       req.left,
			Uploaded:   req.uploaded,
			Event:      req.event,
			IP:         binary.BigEndian.Uint32(ip),
			Port:       req.port,
		}

		res := &announceResponse{
			action:        1,
			transactionID: req.transactionID,
			interval:      5,
			leechers:      0,
			seeders:       0,
			peers:         [][6]byte{},
		}
		pack := res.pack()

		_, err := conn.WriteToUDP(pack, addr)
		if err != nil {
			log.Error().Err(err).Msg("cant write udp")
			return
		}
		return
	}

	state <- eventAnnounce{
		InfoHash:   req.infoHash,
		PeerId:     req.peerID,
		Downloaded: req.downloaded,
		Left:       req.left,
		Uploaded:   req.uploaded,
		Event:      req.event,
		IP:         binary.BigEndian.Uint32(ip),
		Port:       req.port,
	}

	leechers, seeders, _, peers := torrent.state()
	res := &announceResponse{
		action:        1,
		transactionID: req.transactionID,
		interval:      30,
		leechers:      leechers,
		seeders:       seeders,
		peers:         peers,
	}
	pack := res.pack()

	_, err = conn.WriteToUDP(pack, addr)
	if err != nil {
		log.Error().Err(err).Msg("cant write udp")
		return
	}
}
