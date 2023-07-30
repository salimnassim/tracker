package tracker

type ServerConfig struct {
	Address string
	URL     string
	DSN     string
}

func NewServerConfig(address string, url string, dsn string) *ServerConfig {
	return &ServerConfig{
		Address: address,
		URL:     url,
		DSN:     dsn,
	}
}
