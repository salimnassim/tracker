package main

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	conns := tracker.NewStore[uint64, uint64]()
	torrents := tracker.NewStore[tracker.InfoHash, *tracker.Torrent]()
	server := tracker.NewServer(conns, torrents)

	udp_port, err := strconv.Atoi(os.Getenv("BT_UDP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse udp port")
	}
	http_port, err := strconv.Atoi(os.Getenv("BT_HTTP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse http port")
	}

	udp := tracker.NewUDPServer(os.Getenv("BT_UDP_ADDRESS"), udp_port)
	http := tracker.NewHTTPServer(os.Getenv("BT_HTTP_ADDRESS"), http_port)

	err = server.Start([]tracker.Serverer{udp, http})
	if err != nil {
		log.Fatal().Err(err)
	}
}
