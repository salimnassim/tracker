# tracker

This is a BitTorrent tracker that relies on PostgreSQL for persistence. All announced torrents will be tracked automatically. The HTTP server provides a torrent index list with magnet metadata support on the `/` route by default.

## Features
- [x] **HTTP Tracker:** Allows tracking of torrents over HTTP.
- [x] **Compact Peer Format:** Utilizes a compact format for representing peers in the tracker.
- [x] **Scrape:** Supports scraping information from the tracker.
- [ ] **UDP Tracker**
- [ ] **IPv6**

## Environment Variables
- `ADDRESS` (default: `0.0.0.0:9999`): Specifies the address and port for the tracker.
- `ANNOUNCE_URL` (default: `http://localhost:9999/announce`): Used for magnet links in the index view.
- `DSN` (default: `postgres://tracker:tracker@localhost:5432/tracker`): PostgreSQL connection string.
- `TEMPLATE_PATH` (default: `../templates/`): Path to the template files.
- `STATIC_PATH` (default: `../static/`): Path to static files.

## Installation

### Local

1. Run `go mod download && go build -v -o ./dist/tracker ./cmd`.
2. The application can be found under the `dist` directory.

### Docker

1. Configure the environment variables under the `backend` block in `docker-compose.yml`.
2. Run `docker compose up`.