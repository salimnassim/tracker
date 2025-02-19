package tracker

import "context"

type Server struct {
	ctx      context.Context
	conns    Storer[uint64, uint32]
	torrents Storer[[20]byte, *Torrent]
}

func NewServer() *Server {
	return &Server{
		ctx:      context.Background(),
		conns:    NewStore[uint64, uint32](),
		torrents: NewStore[[20]byte, *Torrent](),
	}
}
