package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	actionHandshake = 0
	actionAnnounce  = 1
	actionScrape    = 2
)

type udpServer struct {
	address string
	port    int
}

type errorResponse struct {
	action        uint32
	transactionID uint32
	message       string
}

func (r *errorResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, []byte(r.message))
	writer.Flush()

	return buffer.Bytes()
}

func NewUDPServer(address string, port int) *udpServer {
	return &udpServer{
		address: address,
		port:    port,
	}
}

func (s *udpServer) Address() string {
	return s.address
}

func (s *udpServer) Port() int {
	return s.port
}

func (s *udpServer) Serve(state chan any, conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent], cache Cacher[InfoHash, *Torrent]) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.address, s.port))
	if err != nil {
		state <- err
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		state <- err
		return
	}

	buffer := make([]byte, 256)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			state <- err
			continue
		}

		go handle(conn, remoteAddr, buffer[:n], state, conns, torrents)
	}
}

func handle(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent]) {
	if len(request) < 16 {
		log.Error().Msgf("packet size less than 16")
		return
	}

	connectionID := binary.BigEndian.Uint64(request[0:8])
	action := binary.BigEndian.Uint32(request[8:12])
	transactionID := binary.BigEndian.Uint32(request[12:16])

	log.Debug().
		Uint64("connection_id", connectionID).
		Uint32("action", action).
		Uint32("transaction_id", transactionID).
		Msg("request")

	switch action {
	case actionHandshake:
		handleHandshake(conn, addr, request, state)
	case actionAnnounce:
		if _, ok := conns.Get(connectionID); !ok {
			log.Error().
				Uint64("connection_id", connectionID).
				Msg("connection id not found for announce")

			res := &errorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				log.Error().Err(err).
					Msg("cant write udp handle error")
				return
			}
			return
		}

		handleAnnounce(conn, addr, request, state, torrents)
	case actionScrape:
		if _, ok := conns.Get(connectionID); !ok {
			log.Error().Uint64("connection_id", connectionID).
				Msg("connection id not found for scrape")
			res := &errorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				log.Error().Err(err).
					Msg("cant write udp handle error")
				return
			}
			return
		}

		handleScrape(conn, addr, request, state, torrents)
	default:
		log.Error().
			Uint64("connection_id", connectionID).
			Uint32("action", action).
			Uint32("transaction_id", transactionID).
			Msgf("unknown action")

		res := &errorResponse{
			action:        3,
			transactionID: transactionID,
			message:       "Unknown action",
		}
		pack := res.pack()

		_, err := conn.WriteToUDP(pack, addr)
		if err != nil {
			log.Error().Err(err).
				Msg("cant write udp unknown action error")
			return
		}
		return
	}
}
