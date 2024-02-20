package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TrackerAnnounce = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "announce",
		Help:      "The total number of announces",
	})
	TrackerAnnounceReply = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "announce_reply",
		Help:      "The total number of announce replies",
	})
	TrackerTorrents = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "torrents",
		Help:      "The total number of tracked torrents over time",
	})
)

var (
	TrackerScrape = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "scrape",
		Help:      "The total number of scrapes",
	})
	TrackerScrapeReply = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "scrape_reply",
		Help:      "The total number of scrape replies",
	})
)
