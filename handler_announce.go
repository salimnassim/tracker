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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	// The total number of promAnnounce.
	promAnnounce = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_announce",
		Help: "The total number of announces",
	})
	// The total number of announce replies.
	promAnnounceReply = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_announce_reply",
		Help: "The total number of announce replies",
	})
	// The total number of tracked torrents over time.
	promTracked = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_tracked",
		Help: "The total number of tracked torrents over time",
	})
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

		w.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
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
			log.Info().Msgf("client info hash is not 20 bytes: %s", infoHash)
			failure := ErrorResponse{
				FailureReason: "info_hash is not valid",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		// has to be exactly 20 bytes
		peerID := []byte(query.Get("peer_id"))
		if len(peerID) != 20 {
			log.Info().Msgf("client peer id is not 20 bytes: %s", peerID)
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
			log.Error().Err(err).Msg("cant log in announce")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		promAnnounce.Inc()

		var torrent Torrent
		// get torrent
		torrent, err = server.store.Torrent(ctx, []byte(req.InfoHash))
		if err != nil && err.Error() != "no rows in result set" {
			log.Error().Err(err).Msg("cant query torrent in announce")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// create torrent not found as we track all announced
		if torrent.ID.IsNil() && err.Error() == "no rows in result set" {
			torrent, err = server.store.AddTorrent(ctx, []byte(req.InfoHash))
			promTracked.Inc()
			if err != nil {
				var pgError *pgconn.PgError
				if errors.As(err, &pgError) {
					if pgError.Code == "23505" {
						log.Error().Err(err).Msg("duplicate torrent info_hash on insert in announce")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				log.Error().Err(err).Msg("cant add torrent to store in announce")
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
				log.Error().Err(err).Msg("cant update peer with key in announce")
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
				log.Error().Err(err).Msg("cant upsert peer in announce")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// increment completed by one if event is sent
		if req.Event == "completed" {
			err := server.store.IncrementTorrent(ctx, torrent.ID)
			if err != nil {
				log.Error().Err(err).Msg("cant increment torrent completed in announce")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		peers, err := server.store.Peers(ctx, torrent.ID)
		if err != nil {
			log.Error().Err(err).Msg("cant get peers in announce")
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
			_, err = buffer.Write(pm)
			if err != nil {
				log.Error().Err(err).Msg("cant write marshalled peer to buffer in announce")
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

		promAnnounceReply.Inc()
		replyBencode(w, announce, http.StatusOK)
	}
}
