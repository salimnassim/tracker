package tracker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

type httpServer struct{}

func NewHTTPServer() *httpServer {
	return &httpServer{}
}

func apiTorrentsHandler(torrents TorrentStore) http.HandlerFunc {
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

func listHandler(torrents TorrentStore, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := torrents.ListTorrents(r.Context())
		if err != nil {
			slog.Error("cant list torrents", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		data := newListPageData(result)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			slog.Error("cant render list page", "error", err)
		}
	}
}

func detailHandler(torrents TorrentStore, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash, err := ParseInfoHash(r.PathValue("hash"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		torrent, err := torrents.GetTorrent(r.Context(), hash)
		if err != nil {
			slog.Error("cant get torrent", "info_hash", hash.String(), "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if torrent == nil {
			http.NotFound(w, r)
			return
		}

		showEvent := r.URL.Query().Get("show_event") == "1"
		data := newDetailPageData(hash, torrent, showEvent)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			slog.Error("cant render detail page", "error", err)
		}
	}
}

func (s *httpServer) Serve(ctx context.Context, config *config, state chan any, conns Storer[uint64, time.Time], torrents TorrentStore) {
	base := template.Must(template.New("base").Funcs(templateFuncs).ParseFS(Templates, "templates/layout.html", "templates/components/*.html"))
	listTmpl := template.Must(template.Must(base.Clone()).ParseFS(Templates, "templates/list.html"))
	detailTmpl := template.Must(template.Must(base.Clone()).ParseFS(Templates, "templates/detail.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", listHandler(torrents, listTmpl))
	mux.HandleFunc("GET /torrents/{hash}", detailHandler(torrents, detailTmpl))
	mux.HandleFunc("GET /api/torrents", apiTorrentsHandler(torrents))
	mux.Handle("GET /static/", http.FileServerFS(Static))

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
