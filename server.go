package tracker

import "context"

type EventConnection struct {
	ConnectionID uint64
}

type Server struct {
	ctx      context.Context
	conns    Storer[uint64, uint64]
	torrents Storer[[20]byte, *Torrent]

	state chan any
}

func NewServer(state chan any, conns Storer[uint64, uint64], torrents Storer[[20]byte, *Torrent]) *Server {
	return &Server{
		ctx:      context.Background(),
		conns:    conns,
		torrents: torrents,
		state:    state,
	}
}

func (s *Server) Start(serverer Serverer) {
	serverer.Serve(s.state, s.conns, s.torrents)
}
