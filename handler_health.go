package tracker

import (
	"net/http"

	"github.com/salimnassim/tracker/metric"
)

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric.TrackerHealth.Inc()
		w.WriteHeader(http.StatusOK)
	}
}
