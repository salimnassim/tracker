package tracker

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

func AnnounceHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := r.URL.Query()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error().Err(err).Msg("unable to split host/port in announce")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if ip == "::1" {
			ip = "127.0.0.1"
		}

		if r.Header.Get("X-Forwarded-For") != "" {
			ip = r.Header.Get("X-Forwarded-For")
		}

		port, err := strconv.ParseInt(query.Get("port"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int port %s", query.Get("port"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uploaded, err := strconv.ParseInt(query.Get("uploaded"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int uploaded %s", query.Get("uploaded"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		downloaded, err := strconv.ParseInt(query.Get("downloaded"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int downloaded %s", query.Get("downloaded"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		left, err := strconv.ParseInt(query.Get("left"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int left %s", query.Get("left"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req := AnnounceRequest{
			InfoHash:   fmt.Sprintf("%x", query.Get("info_hash")),
			PeerID:     query.Get("peer_id"),
			Event:      query.Get("event"),
			IP:         ip,
			Port:       int(port),
			Uploaded:   int(uploaded),
			Downloaded: int(downloaded),
			Left:       int(left),
		}

		err = server.validator.Struct(req)
		if err != nil {
			// todo: wrap errors
			errors := err.(validator.ValidationErrors)
			for _, v := range errors {
				log.Error().Msgf("error validation field %s: %s", v.Field(), v.Error())
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var torrent *Torrent
		torrent, err = server.store.GetTorrent(ctx, []byte(req.InfoHash))
		if err != nil && err.Error() != "no rows in result set" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if torrent == nil && err.Error() == "no rows in result set" {
			torrent, err = server.store.AddTorrent(ctx, []byte(req.InfoHash))
			if err != nil {
				var pgError *pgconn.PgError
				if errors.As(err, &pgError) {
					if pgError.Code == "23505" {
						log.Error().Err(err).Msg("duplicate torrent info_hash on insert")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				log.Error().Err(err).Msg("cant add torrente to store")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = server.store.UpsertPeer(ctx, torrent.ID, req)
		if err != nil {
			log.Error().Err(err).Msg("cant add peer to store")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debug().Msgf("%v", torrent)

	}
}
