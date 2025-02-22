package tracker

import "context"

type EventConnection struct {
	ConnectionID uint64
}

type EventAnnounce struct {
	InfoHash   [20]byte
	PeerId     [20]byte
	Downloaded uint64
	Left       uint64
	Uploaded   uint64
	Event      uint32
	IP         uint32
	Port       uint16
}

type EventRegisterTorrent struct {
	InfoHash [20]byte
}

type Server struct {
	ctx      context.Context
	conns    Storer[uint64, uint64]
	torrents Storer[InfoHash, *Torrent]

	state chan any
}

func NewServer(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) *Server {
	return &Server{
		ctx:      context.Background(),
		conns:    conns,
		torrents: torrents,
		state:    state,
	}
}

func (s *Server) StartUDP(serverer UDPServerer) {
	serverer.Serve(s.state, s.conns, s.torrents)
}

func (s *Server) StartHTTP(serverer HTTPServerer) {
	serverer.Serve(s.conns, s.torrents)
}
