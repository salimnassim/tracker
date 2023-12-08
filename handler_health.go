package tracker

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// The total number of health checks.
	promChecks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tracker_health",
		Help: "The total number of health checks",
	})
)

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promChecks.Inc()
		w.WriteHeader(http.StatusOK)
	}
}
