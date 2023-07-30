package tracker

import (
	"time"

	"github.com/gofrs/uuid"
)

type Torrent struct {
	ID        uuid.UUID `db:"id"`
	InfoHash  []byte    `db:"info_hash"`
	Completed int       `db:"completed"`
	CreatedAt time.Time `db:"created_at"`
}
