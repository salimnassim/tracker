package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type HTTPServerer interface {
	Serve(conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent])
}

type HTTPServer struct {
	address string
	port    int
}

func NewHTTPServer(address string, port int) *HTTPServer {
	return &HTTPServer{
		address: address,
		port:    port,
	}
}

func handler(torrents Storer[InfoHash, *Torrent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(torrents)
		if err != nil {
			log.Error().Err(err).Msg("cant marshal torrents")
			return
		}
	}
}

func (s *HTTPServer) Serve(state chan any, conns Storer[uint64, uint64], torrents Storer[InfoHash, *Torrent]) {
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
