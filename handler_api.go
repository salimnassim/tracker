package tracker

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func APIListHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()

		torrents, err := server.store.GetTorrents(ctx)
		if err != nil {
			log.Error().Err(err).Msg("cant query db torrents in api")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(torrents)
		if err != nil {
			log.Error().Err(err).Msg("cant marshal torrents in api")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}
