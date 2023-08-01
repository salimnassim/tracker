package tracker

import (
	"net"
	"testing"
)

func TestPeerMarshal(t *testing.T) {
	peer := &Peer{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 9999,
	}

	want := "\u007f\x00\x00\x01'\x0f"
	got, err := peer.Marshal()
	if err != nil {
		t.Error(err)
	}

	if want != string(got) {
		t.Errorf("want: %v, got %v", want, got)
	}
}

func TestPeerClient(t *testing.T) {
	peer := &Peer{
		PeerID: []byte{45, 84, 82, 51, 48, 48, 48, 45, 100, 121, 98, 119, 54, 108, 115, 110, 115, 99, 49, 55},
	}

	got := peer.Client()
	want := "Transmission"

	if got != want {
		t.Errorf("want: %v, got %v", want, got)
	}

}
