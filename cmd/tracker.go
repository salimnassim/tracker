package main

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	udpPort, err := strconv.Atoi(os.Getenv("BT_UDP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse udp port")
	}
	httpPort, err := strconv.Atoi(os.Getenv("BT_HTTP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse http port")
	}
	peerInterval, err := strconv.Atoi(os.Getenv("PEER_INTERVAL"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse peer interval")
	}
	peerLifetime, err := strconv.Atoi(os.Getenv("PEER_LIFETIME"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse peer lifetime")
	}

	config := tracker.NewConfig(
		os.Getenv("BT_HTTP_ADDRESS"), httpPort,
		os.Getenv("BT_UDP_ADDRESS"), udpPort,
		os.Getenv("BT_UDP_URL"),
		time.Duration(peerInterval), time.Duration(peerLifetime),
	)

	conns := tracker.NewStore[uint64, time.Time]()
	torrents := tracker.NewStore[tracker.InfoHash, *tracker.Torrent]()

	server, err := tracker.NewServer(conns, torrents)
	if err != nil {
		log.Error().Err(err)
		return
	}

	udp := tracker.NewUDPServer()
	http := tracker.NewHTTPServer()

	err = server.Start(config, []tracker.Serverer{udp, http})
	if err != nil {
		log.Error().Err(err)
		return
	}
}
