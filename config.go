package tracker

import "time"

type config struct {
	httpAddress string
	httpPort    int

	udpAddress string
	udpPort    int
	udpURL     string

	cacheInterval time.Duration
	peerInterval  time.Duration
	peerLifetime  time.Duration
}

func NewConfig(httpAddress string, httpPort int, udpAddress string, udpPort int, udpURL string, cacheInterval, peerInterval, peerLifetime time.Duration) *config {
	return &config{
		httpAddress:   httpAddress,
		httpPort:      httpPort,
		udpAddress:    udpAddress,
		udpPort:       udpPort,
		udpURL:        udpURL,
		cacheInterval: cacheInterval,
		peerInterval:  peerInterval,
		peerLifetime:  peerLifetime,
	}
}
