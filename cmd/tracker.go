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

	udpPort, err := strconv.Atoi(os.Getenv("BT_UDP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse udp port")
	}
	httpPort, err := strconv.Atoi(os.Getenv("BT_HTTP_PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse http port")
	}
	_, err = strconv.Atoi(os.Getenv("CACHE_LIFETIME"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant parse cache lifetime")
	}

	conns := tracker.NewStore[uint64, uint64]()
	torrents := tracker.NewStore[tracker.InfoHash, *tracker.Torrent]()
	cache := tracker.NewCache[tracker.InfoHash, *tracker.Torrent]()

	server := tracker.NewServer(conns, torrents, cache)

	udp := tracker.NewUDPServer(os.Getenv("BT_UDP_ADDRESS"), udpPort)
	http := tracker.NewHTTPServer(os.Getenv("BT_HTTP_ADDRESS"), httpPort)

	err = server.Start([]tracker.Serverer{udp, http})
	if err != nil {
		log.Fatal().Err(err)
	}
}
