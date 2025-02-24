package tracker

import (
	"slices"
	"testing"
)

func TestConnectRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		bytes := []byte{0x00, 0x04, 0x17, 0x27, 0x10, 0x19, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x56, 0x78}
		req := &handshakeRequest{}
		err := req.unpack(bytes)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("size", func(t *testing.T) {
		bytes := []byte{0x00, 0x04, 0x17}
		req := &handshakeRequest{}
		err := req.unpack(bytes)

		if err != errorSizeMismatch {
			t.Errorf("should fail on bad size")
		}
	})
}

func TestConnectResponse(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		bytes := []byte{
			0x00, 0x00, 0x00, 0x00,
			0x12, 0x34, 0x56, 0x78,
			0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
		}

		res := &handshakeResponse{
			action:        0,
			transactionID: 0x12345678,
			connectionID:  0x0123456789ABCDEF,
		}
		packed := res.pack()

		if !slices.Equal(bytes, packed) {
			t.Errorf("byte slices are not equal")
		}
	})
}
