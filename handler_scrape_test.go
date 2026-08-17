package tracker

import "testing"

func TestScrapeRequestUnpack(t *testing.T) {
	t.Run("ok single hash", func(t *testing.T) {
		bytes := append(
			[]byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // connection id
				0x00, 0x00, 0x00, 0x02, // action (scrape)
				0x12, 0x34, 0x56, 0x78, // transaction id
			},
			make([]byte, 20)..., // one info hash
		)
		req := &scrapeRequest{}
		if err := req.unpack(bytes); err != nil {
			t.Error(err)
		}
	})

	t.Run("too short", func(t *testing.T) {
		req := &scrapeRequest{}
		if err := req.unpack(make([]byte, 19)); err != errorSizeMismatch {
			t.Errorf("should fail on bad size")
		}
	})

	t.Run("minimum length no hashes triggers panic", func(t *testing.T) {
		req := &scrapeRequest{}
		if err := req.unpack(make([]byte, 20)); err != errorSizeMismatch {
			t.Errorf("should reject non-multiple-of-20 trailing bytes, got %v", err)
		}
	})

	t.Run("misaligned trailing bytes", func(t *testing.T) {
		req := &scrapeRequest{}
		if err := req.unpack(make([]byte, 39)); err != errorSizeMismatch { // 16 + 23, not a multiple of 20
			t.Errorf("should reject misaligned length")
		}
	})
}

func FuzzScrapeUnpack(f *testing.F) {
	f.Add(make([]byte, 16))
	f.Add(make([]byte, 20)) // the panic-triggering length pre-fix
	f.Add(make([]byte, 36)) // 16 + one full hash
	f.Add(make([]byte, 39)) // misaligned

	f.Fuzz(func(t *testing.T, data []byte) {
		req := &scrapeRequest{}
		err := req.unpack(data)
		if err != nil && err != errorSizeMismatch {
			t.Errorf("unpack returned unexpected error: %v", err)
		}
	})
}
