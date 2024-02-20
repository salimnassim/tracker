package tracker

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/tracker/metric"
)

func ScrapeHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		metric.TrackerScrape.Inc()

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

		metric.TrackerScrapeReply.Inc()

		replyBencode(w, scrape, http.StatusOK)
	}
}
