package tracker

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func IndexHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		torrents, err := server.store.AllTorrents(ctx)
		if err != nil {
			log.Error().Err(err).Msg("cant get torrents in index")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(filepath.Join(server.config.TemplatePath, "index.html"))
		if err != nil {
			log.Error().Err(err).Msg("cant parse template in index")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// todo: make a struct for view
		dto := map[string]interface{}{
			"Torrents":    torrents,
			"AnnounceURL": server.config.AnnounceURL,
		}

		err = tmpl.Execute(w, dto)
		if err != nil {
			log.Error().Err(err).Msg("cant execute template in index")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func TorrentHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		uuid, err := uuid.FromString(vars["id"])
		if err != nil {
			log.Error().Err(err).Msg("cant create uuid from string in torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		peers, err := server.store.AllPeers(ctx, uuid)
		if err != nil {
			log.Error().Err(err).Msg("cant get peers in torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(filepath.Join(server.config.TemplatePath, "torrent.html"))
		if err != nil {
			log.Error().Err(err).Msg("cant parse template in torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// todo: make a struct for view
		dto := map[string]interface{}{
			"Peers": peers,
		}

		err = tmpl.Execute(w, dto)
		if err != nil {
			log.Error().Err(err).Msg("cant execute template in torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
