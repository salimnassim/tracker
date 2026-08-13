package tracker

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type httpServer struct{}

func NewHTTPServer() *httpServer {
	return &httpServer{}
}

func handler(torrents TorrentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := torrents.ListTorrents(r.Context())
		if err != nil {
			slog.Error("cant list torrents", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			slog.Error("cant marshal torrents", "error", err)
			return
		}
	}
}

func (s *httpServer) Serve(config *config, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(torrents))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.httpAddress, config.httpPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("http server error", "error", err)
	}
}
