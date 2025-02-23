package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

type Serverer interface {
	Serve(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent])
	Address() string
	Port() int
}

type UDPServer struct {
	address string
	port    int
}

type ErrorResponse struct {
	action        uint32
	transactionID uint32
	message       string
}

func (r *ErrorResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, []byte(r.message))
	writer.Flush()

	return buffer.Bytes()
}

func NewUDPServer(address string, port int) *UDPServer {
	return &UDPServer{
		address: address,
		port:    port,
	}
}

func (s *UDPServer) Address() string {
	return s.address
}

func (s *UDPServer) Port() int {
	return s.port
}

func (s *UDPServer) Serve(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) {
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

func handle(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) {
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
	case 0:
		handleHandshake(conn, addr, request, state)
	case 1:
		if _, ok := conns.Get(connectionID); !ok {
			log.Error().Uint64("connection_id", connectionID).Msg("connection id not found for announce")
			res := &ErrorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				log.Error().Err(err).Msg("cant write udp handle error")
				return
			}
			return
		}

		handleAnnounce(conn, addr, request, state, torrents)
	case 2:
		if _, ok := conns.Get(connectionID); !ok {
			log.Error().Uint64("connection_id", connectionID).Msg("connection id not found for scrape")
			res := &ErrorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				log.Error().Err(err).Msg("cant write udp handle error")
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

		res := &ErrorResponse{
			action:        3,
			transactionID: transactionID,
			message:       "Unknown action",
		}
		pack := res.pack()

		_, err := conn.WriteToUDP(pack, addr)
		if err != nil {
			log.Error().Err(err).Msg("cant write udp handshake error")
			return
		}
		return
	}
}
