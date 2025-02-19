package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	errorSizeMismatch = errors.New("bad message size")
)

type connectRequest struct {
	protocolID    uint64 // 0x41727101980
	action        uint32
	transactionID uint32
}

func (r *connectRequest) unpack(bytes []byte) error {
	if len(bytes) != 16 {
		return errorSizeMismatch
	}

	r.protocolID = binary.BigEndian.Uint64(bytes[0:8])
	r.action = binary.BigEndian.Uint32(bytes[8:12])
	r.transactionID = binary.BigEndian.Uint32(bytes[12:16])

	return nil
}

type connectReply struct {
	action        uint32
	transactionID uint32
	connectionID  uint64
}

func (r *connectReply) pack() []byte {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_ = binary.Write(writer, binary.BigEndian, r.action)
	_ = binary.Write(writer, binary.BigEndian, r.transactionID)
	_ = binary.Write(writer, binary.BigEndian, r.connectionID)
	writer.Flush()

	return buffer.Bytes()
}
