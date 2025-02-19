package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

type packer interface {
	pack() []byte
}

type unpacker interface {
	unpack()
}

type errorResponse struct {
	// should be 3 always
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

type connectResponse struct {
	action        uint32
	transactionID uint32
	connectionID  uint64
}

func (r *connectResponse) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, r.connectionID)
	writer.Flush()

	return buffer.Bytes()
}

type announceResponse struct {
	action        uint32
	transactionID uint32
	interval      uint32
	leechers      uint32
	seeders       uint32
	peers         []byte
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
