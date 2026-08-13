# tracker

A UDP BitTorrent tracker written in Go. It handles the connect/announce/scrape handshake, persists torrents and peers to SQLite, and exposes a small HTTP endpoint for inspecting swarm state.

## Contents

- `server_udp.go`, `handler_handshake.go`, `handler_announce.go`, `handler_scrape.go` the UDP tracker protocol: connection ID handshake, peer announces, and swarm scraping.
- `server_http.go` a read-only HTTP endpoint (`GET /`) that dumps all known torrents and peers as JSON.
- `server.go` the event loop tying UDP/HTTP servers together, applying announces, registering new torrents, and expiring stale peers on a timer.
- `torrent.go`, `peer.go` core domain types (`Torrent`, `Peer`, `InfoHash`, `PeerID`) and swarm state (seeders/leechers/completed) derivation.
- `store.go` a generic in-memory, JSON-marshalable key/value store used for tracking active UDP connection IDs.
- `db/` SQLite-backed `TorrentStore` implementation.
- `cmd/tracker.go` the binary entrypoint: reads configuration from the environment, opens the database, runs migrations, and starts the server.

## Install

```sh
go build -o ./tracker ./cmd
```

or with Docker:

```sh
docker compose up --build
```

## Usage

The tracker is configured entirely through environment variables:

| Variable          | Description                                       | Example                    |
| ----------------- | -------------------------------------------------- | --------------------------- |
| `BT_HTTP_ADDRESS`  | Address the HTTP stats server binds to             | `` (all interfaces)         |
| `BT_HTTP_PORT`     | Port for the HTTP stats server                     | `8080`                       |
| `BT_UDP_ADDRESS`   | Address the UDP tracker binds to                   | `` (all interfaces)         |
| `BT_UDP_PORT`      | Port for the UDP tracker                           | `6881`                       |
| `BT_UDP_URL`       | Announce URL embedded in generated magnet links    | `udp://localhost:6881`      |
| `PEER_INTERVAL`    | Seconds a client is told to wait between announces | `1200`                       |
| `PEER_LIFETIME`    | Seconds of inactivity before a peer is expired      | `3600`                       |
| `TRACKER_DB_PATH`  | Path to the SQLite database file                   | `./tracker.db`              |

```sh
BT_HTTP_ADDRESS= BT_HTTP_PORT=8080 \
BT_UDP_ADDRESS= BT_UDP_PORT=6881 BT_UDP_URL=udp://localhost:6881 \
PEER_INTERVAL=1200 PEER_LIFETIME=3600 \
TRACKER_DB_PATH=./tracker.db \
./tracker
```

Torrents are registered automatically on their first announce, there is no separate registration step or tracker admin API. Point a client at:

```
udp://<host>:<udp-port>
```

and check swarm state at `http://<host>:<http-port>/`.

## Development

```sh
make test           # go test -v ./...
make build           # CGO_ENABLED=0 go build -o ./tracker ./cmd
make sqlc-generate    # regenerate db/*.sql.go from queries/*.sql via sqlc
make ci              # fmt, vet, race tests, mod tidy check, govulncheck
```

Database code is generated, not hand-written: change `queries/*.sql` or add a migration under `migrations/`, then run `make sqlc-generate`.