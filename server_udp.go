package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const (
	actionHandshake = 0
	actionAnnounce  = 1
	actionScrape    = 2
)

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

type udpServer struct{}

func NewUDPServer() *udpServer {
	return &udpServer{}
}

func (s *udpServer) Serve(config *config, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.udpAddress, config.udpPort))
	if err != nil {
		state <- err
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		state <- err
		return
	}

	for {
		buffer := make([]byte, 256)
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			state <- err
			continue
		}

		go handle(conn, remoteAddr, buffer[:n], state, conns, torrents)
	}
}

func handle(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore) {
	if len(request) < 16 {
		slog.Error("packet size less than 16")
		return
	}

	connectionID := binary.BigEndian.Uint64(request[0:8])
	action := binary.BigEndian.Uint32(request[8:12])
	transactionID := binary.BigEndian.Uint32(request[12:16])

	slog.Debug("request",
		"connection_id", connectionID,
		"action", action,
		"transaction_id", transactionID)

	switch action {
	case actionHandshake:
		handleHandshake(conn, addr, request, state)
	case actionAnnounce:
		if _, ok := conns.Get(connectionID); !ok {
			slog.Error("connection id not found for announce", "connection_id", connectionID)

			res := &errorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				slog.Error("cant write udp handle error", "error", err)
				return
			}
			return
		}

		handleAnnounce(conn, addr, request, state, torrents)
	case actionScrape:
		if _, ok := conns.Get(connectionID); !ok {
			slog.Error("connection id not found for scrape", "connection_id", connectionID)
			res := &errorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "Invalid connection ID",
			}
			pack := res.pack()

			_, err := conn.WriteToUDP(pack, addr)
			if err != nil {
				slog.Error("cant write udp handle error", "error", err)
				return
			}
			return
		}

		handleScrape(conn, addr, request, state, torrents)
	default:
		slog.Error("unknown action",
			"connection_id", connectionID,
			"action", action,
			"transaction_id", transactionID)

		res := &errorResponse{
			action:        3,
			transactionID: transactionID,
			message:       "Unknown action",
		}
		pack := res.pack()

		_, err := conn.WriteToUDP(pack, addr)
		if err != nil {
			slog.Error("cant write udp unknown action error", "error", err)
			return
		}
		return
	}
}
