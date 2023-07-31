package tracker

import (
	"bytes"
	"errors"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cristalhq/bencode"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

func AnnounceHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo: middleware
		w.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")

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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		uploaded, err := strconv.ParseInt(query.Get("uploaded"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int uploaded %s", query.Get("uploaded"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		downloaded, err := strconv.ParseInt(query.Get("downloaded"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int downloaded %s", query.Get("downloaded"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		left, err := strconv.ParseInt(query.Get("left"), 10, 0)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse int left %s", query.Get("left"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// some magnet downloads report left as maxint, default it to one
		if left == math.MaxInt {
			left = 1
		}

		// has to be exactly 20 bytes
		infoHash := []byte(query.Get("info_hash"))
		if len(infoHash) != 20 {
			log.Info().Msgf("client info hash is not 20 bytes: %s", infoHash)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// has to be exactly 20 bytes
		peerID := []byte(query.Get("peer_id"))
		if len(peerID) != 20 {
			log.Info().Msgf("client peer id is not 20 bytes: %s", peerID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req := AnnounceRequest{
			InfoHash:   infoHash,
			PeerID:     peerID,
			Event:      query.Get("event"),
			IP:         ip,
			Port:       int(port),
			Key:        query.Get("key"),
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

		// get torrent
		var torrent Torrent
		torrent, err = server.store.GetTorrent(ctx, []byte(req.InfoHash))
		if err != nil && err.Error() != "no rows in result set" {
			log.Error().Err(err).Msg("cant query torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// create torrent not found as we track all announced
		if torrent.ID.IsNil() && err.Error() == "no rows in result set" {
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
				log.Error().Err(err).Msg("cant add torrent to store")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// try to update existing record by using query string key

		// ok is true if peer was updated with a key
		var ok bool
		if query.Get("key") != "" {
			err, ok = server.store.UpdatePeerWithKey(ctx, torrent.ID, req)
			if err != nil {
				log.Error().Err(err).Msg("cant update peer with key in store")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if !ok {
			err = server.store.InsertOrUpdatePeer(ctx, torrent.ID, req)
			if err != nil {
				log.Error().Err(err).Msg("cant update peer in store")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		peers, err := server.store.GetPeers(ctx, torrent.ID)
		if err != nil {
			log.Error().Err(err).Msg("cant get peers from store")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		buffer := new(bytes.Buffer)
		for _, p := range peers {
			pm, err := p.Marshal()
			if err != nil {
				log.Error().Err(err).Msg("cant marshal peer in announce")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			buffer.Write(pm)
		}

		// todo: make a struct for response
		announce := map[string]interface{}{
			"interval":     60 * time.Second,
			"min interval": 120 * time.Second,
			"complete":     torrent.Seeders,
			"incomplete":   torrent.Leechers,
			"peers":        buffer.String(),
		}

		res, err := bencode.Marshal(announce)
		if err != nil {
			log.Error().Err(err).Msg("cant bencode in announce")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(res)
	}
}
