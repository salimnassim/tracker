package tracker

import "encoding/binary"

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

func (r *announceRequest) unpack(bytes []byte) error {
	if len(bytes) != 98 {
		return errorSizeMismatch
	}

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

	return nil
}
