package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// The total number of health checks.
	TrackerHealth = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "tracker",
		Name:      "health",
		Help:      "The total number of health checks",
	})
)
