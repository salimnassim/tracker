package tracker

import "time"

type config struct {
	httpAddress string
	httpPort    int

	udpAddress string
	udpPort    int
	udpURL     string

	peerInterval time.Duration
	peerLifetime time.Duration

	dbPath string
}

func NewConfig(httpAddress string, httpPort int, udpAddress string, udpPort int, udpURL string, peerInterval, peerLifetime time.Duration, dbPath string) *config {
	return &config{
		httpAddress:  httpAddress,
		httpPort:     httpPort,
		udpAddress:   udpAddress,
		udpPort:      udpPort,
		udpURL:       udpURL,
		peerInterval: peerInterval,
		peerLifetime: peerLifetime,
		dbPath:       dbPath,
	}
}
