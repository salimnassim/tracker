package tracker

import (
	"context"
	"html/template"
	"os"
	"path"
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
	templates Templater
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
		templates: NewTemplateStore(),
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

func (sv *Server) CacheTemplates() {
	// index
	tplIndex := template.Must(
		template.ParseFiles(
			path.Join(os.Getenv("TEMPLATE_PATH"), "index.html"),
		),
	)
	sv.templates.Add(TemplateIndex, tplIndex)
	// torrent
	tplTorrent := template.Must(
		template.ParseFiles(
			path.Join(os.Getenv("TEMPLATE_PATH"), "torrent.html"),
		),
	)
	sv.templates.Add(TemplateTorrent, tplTorrent)
}
