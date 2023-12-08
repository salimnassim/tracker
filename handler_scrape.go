package tracker

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	// The total number of promAdded torrents.
	promScrape = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_scrape",
		Help: "The total number of scrapes",
	})
	promScrapeReply = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_scrape_reply",
		Help: "The total number of scrape replies",
	})
)

func ScrapeHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")

		promScrape.Inc()

		infoHash, ok := r.URL.Query()["info_hash"]
		if !ok {
			log.Error().Msg("scrape info_hash is not present")
			failure := ErrorResponse{
				FailureReason: "info_hash is not present",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		var hashes [][]byte
		for _, v := range infoHash {
			hashes = append(hashes, []byte(v))
		}

		torrents, err := server.store.Scrape(ctx, hashes)
		if err != nil {
			log.Error().Err(err).Msg("unable to fetch torrents in scrape")
			failure := ErrorResponse{
				FailureReason: "internal server error",
			}
			replyBencode(w, failure, http.StatusBadRequest)
			return
		}

		scrape := ScrapeResponse{
			Files: make(map[string]ScrapeTorrent),
		}
		for _, t := range torrents {
			scrape.Files[string(t.InfoHash)] = ScrapeTorrent{
				Complete:   t.Seeders,
				Incomplete: t.Leechers,
				Downloaded: t.Completed,
			}
		}

		promScrapeReply.Inc()

		replyBencode(w, scrape, http.StatusOK)
	}
}
