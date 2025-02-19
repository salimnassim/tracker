package tracker

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand/v2"
	"net"

	"github.com/rs/zerolog/log"
)

const (
	ActionAnnounce = 1
	ActionScrape   = 2
)

type UDPServer struct {
	address  string
	port     int
	conns    Storer[uint64, uint32]
	torrents Storer[[20]byte, *Torrent]
}

func NewUDPServer(address string, port int, conns Storer[uint64, uint32], torrents Storer[[20]byte, *Torrent]) *UDPServer {
	return &UDPServer{
		address:  address,
		port:     port,
		conns:    conns,
		torrents: torrents,
	}
}

func (s *UDPServer) Serve(errs chan error) {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8888")
	if err != nil {
		errs <- err
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		errs <- err
		return
	}

	for {
		buffer := make([]byte, 256)
		_, remoteAddr, err := conn.ReadFrom(buffer)

		if err != nil {
			log.Error().Err(err).Msg("cant read from udp buffer")
			errs <- err
		}

		log.Debug().Msgf("%v", buffer)

		// check for magic constant, create connectionID and store it for the future
		if binary.BigEndian.Uint64(buffer[0:8]) == 0x41727101980 {
			action := binary.BigEndian.Uint32(buffer[8:12])

			// action has to be announce
			if action != 0 {
				log.Error().Uint32("action", action).Msg("magic action is not announce")
				continue
			}

			transactionID := binary.BigEndian.Uint32(buffer[12:16])
			connectionID := rand.Uint64()

			log.Info().Uint64("connection_id", connectionID).Uint32("transaction_id", transactionID).Msg("new connection")
			s.conns.Set(connectionID, transactionID)

			res := &connectResponse{
				action:        0,
				transactionID: transactionID,
				connectionID:  connectionID,
			}

			packed := res.pack()
			_, err = conn.WriteTo(packed, remoteAddr)
			if err != nil {
				log.Error().Err(err).Msg("cant write to connection")
				errs <- err
			}
			continue
		}

		connectionID := binary.BigEndian.Uint64(buffer[0:8])
		transactionID := binary.BigEndian.Uint32(buffer[12:16])
		action := binary.BigEndian.Uint32(buffer[8:12])

		if _, ok := s.conns.Get(connectionID); !ok {
			log.Error().Uint64("info_hash", connectionID).Uint32("action", action).Msg("no connection id")

			res := errorResponse{
				action:        3,
				transactionID: transactionID,
				message:       "no connection id",
			}
			packed := res.pack()
			_, err = conn.WriteTo(packed, remoteAddr)
			if err != nil {
				log.Error().Err(err).Msg("cant write to connection")
				errs <- err
			}
			continue
		}

		// announce
		if action == ActionAnnounce {
			req := announceRequest{}
			req.unpack(buffer)

			torrent, ok := s.torrents.Get(req.infoHash)
			if !ok {
				copy := &Torrent{
					Peers: []Peer{
						{
							req.ip, req.port,
						},
					},
				}
				torrent = copy
				s.torrents.Set(req.infoHash, copy)

				log.Info().Str("info_hash", hex.EncodeToString(req.infoHash[:])).Msg("create torrent")
			}

			log.Info().Str("info_hash", hex.EncodeToString(req.infoHash[:])).Msg("announce torrent")

			peers := torrent.pack()
			res := announceResponse{
				action:        1,
				transactionID: 0,
				interval:      5,
				leechers:      0,
				seeders:       99,
				peers:         peers,
			}

			packed := res.pack()
			_, err = conn.WriteTo(packed, remoteAddr)
			if err != nil {
				log.Error().Err(err).Msg("cant write to connection")
				errs <- err
			}
			continue
		}

		if action == ActionScrape {
			// req := scrape{}
			// req.unpack(buffer)

			// log.Debug().Msgf("%+v", req)
		}
	}
}
