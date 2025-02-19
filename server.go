package tracker

import "context"

type Server struct {
	ctx      context.Context
	conns    Storer[uint64, uint64]
	torrents Storer[[20]byte, *Torrent]

	state  chan any
	errors chan error
	stop   chan bool
}

func NewServer(errors chan error, stop chan bool, state chan any) *Server {
	return &Server{
		ctx:      context.Background(),
		conns:    NewStore[uint64, uint64](),
		torrents: NewStore[[20]byte, *Torrent](),
		state:    state,
		errors:   errors,
		stop:     stop,
	}
}

func (s *Server) Start(serverer Serverer) {
	serverer.Serve(s.state, s.errors, s.stop, s.conns, s.torrents)
}
