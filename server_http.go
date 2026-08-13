package tracker

import (
	"context"
	"encoding/json"
	"errors"
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

func (s *httpServer) Serve(ctx context.Context, config *config, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(torrents))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.httpAddress, config.httpPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown error", "error", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("http server error", "error", err)
	}
}
