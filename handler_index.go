package tracker

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func HandlerIndex(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		torrents, err := server.store.GetTorrents(ctx)
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
