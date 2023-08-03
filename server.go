package tracker

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config    *ServerConfig
	validator *validator.Validate
	pool      *pgxpool.Pool
	store     TorrentStorable
}

func NewServer(config *ServerConfig) *Server {
	pgxconfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse pgxpool config")
	}
	pgxpool, err := pgxpool.NewWithConfig(context.Background(), pgxconfig)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create pgxpool")
	}

	return &Server{
		config:    config,
		validator: validator.New(),
		pool:      pgxpool,
		store:     NewTorrentStore(pgxpool),
	}
}

// Creates a goroutine that runs function f every duration d.
// ts gives access to the store.
func (sv *Server) RunTask(d time.Duration, f func(ts TorrentStorable)) {
	go func() {
		for range time.Tick(d) {
			f(sv.store)
		}
	}()
}
