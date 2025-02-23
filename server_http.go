package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type httpServer struct {
	address string
	port    int
}

func NewHTTPServer(address string, port int) *httpServer {
	return &httpServer{
		address: address,
		port:    port,
	}
}

func (s *httpServer) Address() string {
	return s.address
}

func (s *httpServer) Port() int {
	return s.port
}

func handler(torrents Storer[InfoHash, *Torrent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// todo: cache

		err := json.NewEncoder(w).Encode(torrents)
		if err != nil {
			log.Error().Err(err).Msg("cant marshal torrents")
			return
		}
	}
}

func (s *httpServer) Serve(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(torrents))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.address, s.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("http server error")
	}
}
