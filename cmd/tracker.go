package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/salimnassim/tracker"
	trackerdb "github.com/salimnassim/tracker/db"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	udpPort, err := strconv.Atoi(os.Getenv("BT_UDP_PORT"))
	if err != nil {
		slog.Error("cant parse udp port", "error", err)
		os.Exit(1)
	}
	httpPort, err := strconv.Atoi(os.Getenv("BT_HTTP_PORT"))
	if err != nil {
		slog.Error("cant parse http port", "error", err)
		os.Exit(1)
	}
	peerInterval, err := strconv.Atoi(os.Getenv("PEER_INTERVAL"))
	if err != nil {
		slog.Error("cant parse peer interval", "error", err)
		os.Exit(1)
	}
	peerLifetime, err := strconv.Atoi(os.Getenv("PEER_LIFETIME"))
	if err != nil {
		slog.Error("cant parse peer lifetime", "error", err)
		os.Exit(1)
	}

	dbPath := os.Getenv("TRACKER_DB_PATH")
	if dbPath == "" {
		dbPath = "./tracker.db"
	}

	config := tracker.NewConfig(
		os.Getenv("BT_HTTP_ADDRESS"), httpPort,
		os.Getenv("BT_UDP_ADDRESS"), udpPort,
		os.Getenv("BT_UDP_URL"),
		time.Duration(peerInterval), time.Duration(peerLifetime),
		dbPath,
	)

	sqlDB, err := trackerdb.Open(dbPath)
	if err != nil {
		slog.Error("cant open database", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(tracker.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("cant set goose dialect", "error", err)
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		slog.Error("cant run migrations", "error", err)
		os.Exit(1)
	}

	conns := tracker.NewStore[uint64, time.Time]()
	torrents := trackerdb.NewTorrentStore(sqlDB)
	defer torrents.Close()

	server, err := tracker.NewServer(ctx, conns, torrents)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		return
	}

	udp := tracker.NewUDPServer()
	http := tracker.NewHTTPServer()

	err = server.Start(config, []tracker.Serverer{udp, http})
	if err != nil {
		slog.Error("server exited", "error", err)
		return
	}
}
