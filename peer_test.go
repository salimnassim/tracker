package tracker

import "testing"

func TestPeerMarshal(t *testing.T) {
	peer := &Peer{
		IP:   "127.0.0.1",
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
