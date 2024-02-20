package tracker

import (
	"bytes"
	"errors"
	"math"
	"net"
	"net/http"
	"slices"
	"strconv"

	"github.com/cristalhq/bencode"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker/metric"
)

// Writes statusCode header and bencoded v.
func replyBencode(w http.ResponseWriter, v any, statusCode int) {
	bytes, err := bencode.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("cant bencode ok reply")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func AnnounceHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := r.URL.Query()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error().Err(err).Str("source", "http_announce").Msg("cant split host port")
			failure := ErrorResponse{
				FailureReason: "internal server error",
			}
			replyBencode(w, failure, http.StatusInternalServerError)
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
			failure := ErrorResponse{
				FailureReason: "port is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		uploaded, err := strconv.ParseInt(query.Get("uploaded"), 10, 0)
		if err != nil {
			failure := ErrorResponse{
				FailureReason: "uploaded is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		downloaded, err := strconv.ParseInt(query.Get("downloaded"), 10, 0)
		if err != nil {
			failure := ErrorResponse{
				FailureReason: "downloaded is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		left, err := strconv.ParseInt(query.Get("left"), 10, 0)
		if err != nil {
			failure := ErrorResponse{
				FailureReason: "left is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		if !slices.Contains([]string{"started", "stopped", "completed", ""}, query.Get("event")) {
			failure := ErrorResponse{
				FailureReason: "event is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		// some magnet downloads report left as maxint, default it to one
		if left == math.MaxInt {
			left = 1
		}

		// has to be exactly 20 bytes
		infoHash := []byte(query.Get("info_hash"))
		if len(infoHash) != 20 {
			log.Info().Str("source", "http_announce").Msgf("client info hash is not 20 bytes: %s", infoHash)
			failure := ErrorResponse{
				FailureReason: "info_hash is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		// has to be exactly 20 bytes
		peerID := []byte(query.Get("peer_id"))
		if len(peerID) != 20 {
			log.Info().Str("source", "http_announce").Msgf("client peer id is not 20 bytes: %s", peerID)
			failure := ErrorResponse{
				FailureReason: "peer_id is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
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
			errors := err.(validator.ValidationErrors)
			for _, v := range errors {
				// send first error as a failure reason
				// todo: make the message more user friendly
				failure := ErrorResponse{
					FailureReason: v.Error(),
				}
				replyBencode(w, failure, http.StatusBadRequest)
				return
			}
			return
		}

		err = server.store.Log(ctx, req)
		if err != nil {
			log.Error().Err(err).Str("source", "http_announce").Msg("cant insert announce log")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		metric.TrackerAnnounce.Inc()

		var torrent Torrent

		// get torrent
		torrent, err = server.store.Torrent(ctx, []byte(req.InfoHash))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error().Err(err).Str("source", "http_announce").Msg("cant get torrent")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// create torrent not found as we track all announced
		if torrent.ID.IsNil() && errors.Is(err, pgx.ErrNoRows) {
			torrent, err = server.store.AddTorrent(ctx, []byte(req.InfoHash))

			metric.TrackerTorrents.Inc()
			if err != nil {
				var pgError *pgconn.PgError
				if errors.As(err, &pgError) {
					if pgError.Code == "23505" {
						log.Error().Err(err).Str("source", "http_announce").Msg("duplicate torrent on insert")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				log.Error().Err(err).Str("source", "http_announce").Msg("cant add torrent to store")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// try to update existing record by using query string key
		// ok is true if peer was updated with a key
		var ok bool
		if query.Get("key") != "" {
			ok, err = server.store.UpdatePeerWithKey(ctx, torrent.ID, req)
			if err != nil {
				log.Error().Err(err).Str("source", "http_announce").Msg("cant update peer with key")
				failure := ErrorResponse{
					FailureReason: "key is not valid",
				}
				replyBencode(w, failure, http.StatusUnauthorized)
				return
			}
		}

		if !ok {
			err = server.store.UpsertPeer(ctx, torrent.ID, req)
			if err != nil {
				log.Error().Err(err).Str("source", "http_announce").Msg("cant upsert peer")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// increment completed by one if event is sent
		if req.Event == "completed" {
			err := server.store.IncrementTorrent(ctx, torrent.ID)
			if err != nil {
				log.Error().Err(err).Str("source", "http_announce").Msg("cant increment torrent completed")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		peers, err := server.store.Peers(ctx, torrent.ID)
		if err != nil {
			log.Error().Err(err).Str("source", "http_announce").Msg("cant get peers")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		buffer := new(bytes.Buffer)
		for _, p := range peers {
			pm, err := p.Marshal()
			if err != nil {
				log.Error().Err(err).Str("source", "http_announce").Msg("cant marshal peer")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = buffer.Write(pm)
			if err != nil {
				log.Error().Err(err).Str("source", "http_announce").Msg("cant write marshalled peer to buffer")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		announce := AnnounceResponse{
			Interval:    60,
			MinInterval: 120,
			Complete:    torrent.Seeders,
			Incomplete:  torrent.Leechers,
			Peers:       buffer.String(),
		}

		metric.TrackerAnnounceReply.Inc()
		replyBencode(w, announce, http.StatusOK)
	}
}
