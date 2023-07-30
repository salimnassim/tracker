package tracker

import (
	"context"

	"github.com/go-playground/validator/v10"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	pgx "github.com/jackc/pgx/v5"
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
	pgxconfig.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		pgxuuid.Register(c.TypeMap())
		return nil
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
