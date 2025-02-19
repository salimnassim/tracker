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

type connectRequest struct {
	protocolID    uint64
	action        uint32
	transactionID uint32
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

type announceRequest struct {
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

func (r *announceRequest) unpack(bytes []byte) {
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
