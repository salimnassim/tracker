package tracker

type ServerConfig struct {
	Address      string
	AnnounceURL  string
	DSN          string
	TemplatePath string
}

func NewServerConfig(address string, announceURL string, dsn string, templatePath string) *ServerConfig {
	return &ServerConfig{
		Address:      address,
		AnnounceURL:  announceURL,
		DSN:          dsn,
		TemplatePath: templatePath,
	}
}
