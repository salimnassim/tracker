package tracker

type AnnounceRequest struct {
	InfoHash string `db:"info_hash" validate:"required,ascii,len=40"`
	PeerID   string `db:"peer_id" validate:"required,ascii,len=20"`
	Event    string `db:"event" validate:"ascii"`
	IP       string `db:"ip" validate:"required,ip"`
	Port     int    `db:"port" validate:"required,number"`
	Key      string `db:"key" validate:"ascii"`

	Uploaded   int `db:"uploaded" validate:"number"`
	Downloaded int `db:"downloaded" validate:"number"`
	Left       int `db:"left" validate:"number"`
}
