package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand/v2"
	"net"

	"github.com/rs/zerolog/log"
)

var errorSizeMismatch = errors.New("bad message size")

type handshakeRequest struct {
	protocolID    uint64 // 0x41727101980
	action        uint32
	transactionID uint32
}

func (r *handshakeRequest) unpack(bytes []byte) error {
	if len(bytes) != 16 {
		return errorSizeMismatch
	}

	r.protocolID = binary.BigEndian.Uint64(bytes[0:8])
	r.action = binary.BigEndian.Uint32(bytes[8:12])
	r.transactionID = binary.BigEndian.Uint32(bytes[12:16])

	return nil
}

type handshakeResponse struct {
	action        uint32
	transactionID uint32
	connectionID  uint64
}

func (r *handshakeResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, r.connectionID)
	writer.Flush()

	return buffer.Bytes()
}

func handleHandshake(conn *net.UDPConn, addr *net.UDPAddr, request []byte, state chan any) {
	req := &handshakeRequest{}
	err := req.unpack(request)
	if err != nil {
		connectionID := binary.BigEndian.Uint64(request[0:8])
		log.Error().Err(err).
			Int64("connection_id", int64(connectionID)).
			Int("size", len(request)).
			Msg("cant unpack handshake request")
		return
	}

	if req.protocolID != 0x41727101980 {
		connectionID := binary.BigEndian.Uint64(request[0:8])
		log.Error().Err(err).
			Int64("connection_id", int64(connectionID)).
			Msg("protocol id is not 0x41727101980")

		res := &errorResponse{
			action:        3,
			transactionID: req.transactionID,
			message:       "Invalid request",
		}
		pack := res.pack()

		_, err = conn.WriteToUDP(pack, addr)
		if err != nil {
			log.Error().Err(err).Msg("cant write udp handshake error")
			return
		}
		return
	}

	connectionID := rand.Uint64()
	state <- eventConnection{
		ConnectionID: connectionID,
	}

	res := &handshakeResponse{
		action:        0,
		transactionID: req.transactionID,
		connectionID:  connectionID,
	}
	pack := res.pack()

	_, err = conn.WriteToUDP(pack, addr)
	if err != nil {
		log.Error().Err(err).Msg("cant write to udp")
		return
	}
}
