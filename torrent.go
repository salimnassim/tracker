package tracker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Torrent struct {
	ID        uuid.UUID `db:"id" json:"id"`
	InfoHash  []byte    `db:"info_hash" json:"info_hash"`
	Completed int       `db:"completed" json:"completed"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	Seeders  int
	Leechers int
}

func (t *Torrent) MarshalJSON() ([]byte, error) {
	type dto Torrent
	return json.Marshal(struct {
		ID        string    `json:"id"`
		InfoHash  string    `json:"info_hash"`
		Completed int       `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
		Seeders   int       `json:"seeders"`
		Leechers  int       `json:"leechers"`
		*dto
	}{
		ID:        t.ID.String(),
		InfoHash:  fmt.Sprintf("%x", string(t.InfoHash)),
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
		Seeders:   t.Seeders,
		Leechers:  t.Leechers,
	})
}
