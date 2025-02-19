package tracker

import (
	"encoding/binary"
	"net"

	"github.com/rs/zerolog/log"
)

const (
	ActionAnnounce = 1
	ActionScrape   = 2
)

type Serverer interface {
	Serve(state chan any, errs chan error, stop chan bool, conns Storer[uint64, uint64], torrents Storer[[20]byte, *Torrent])
}

type UDPServer struct {
	address string
	port    int
}

func NewUDPServer(address string, port int) *UDPServer {
	return &UDPServer{
		address: address,
		port:    port,
	}
}

func (s *UDPServer) Serve(state chan any, errs chan error, stop chan bool, conns Storer[uint64, uint64], torrents Storer[[20]byte, *Torrent]) {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8888")
	if err != nil {
		errs <- err
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		errs <- err
		return
	}

	buffer := make([]byte, 256)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)

		log.Debug().Msg("message read")

		if err != nil {
			log.Error().Err(err)
			errs <- err
			continue
		}

		go handle(conn, remoteAddr, buffer[:n], state)
	}
}

func handle(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any) {
	if len(request) < 16 {
		log.Error().Msgf("packet size less than 16")
		return
	}

	connectionID := binary.BigEndian.Uint64(request[0:8])
	action := binary.BigEndian.Uint32(request[8:12])
	transactionID := binary.BigEndian.Uint32(request[12:16])

	switch action {
	case 0:
		handleConnection(conn, addr, request, state)

	// case 1:
	// 	handleAnnounce(conn, addr, connectionID, transactionID, request)

	// case 2:
	// 	handleAnnounce(conn, addr, connectionID, transactionID, request)

	default:
		log.Error().
			Uint64("connection_id", connectionID).
			Uint32("action", action).
			Uint32("transaction_id", transactionID).
			Msgf("unknown action")
	}
}
