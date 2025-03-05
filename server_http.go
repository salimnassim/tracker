package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type httpServer struct{}

func NewHTTPServer() *httpServer {
	return &httpServer{}
}

func handler(cache Cacher[InfoHash, *Torrent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(cache)
		if err != nil {
			log.Error().Err(err).Msg("cant marshal torrents")
			return
		}
	}
}

func (s *httpServer) Serve(config *config, state chan any, conns Storer[uint64, time.Time], torrents Storer[InfoHash, *Torrent], cache Cacher[InfoHash, *Torrent]) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(cache))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.httpAddress, config.httpPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("http server error")
	}
}
