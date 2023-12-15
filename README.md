# tracker
BitTorrent tracker which relies on PostgreSQL for persistence. All announced torrents will be tracked automatically.
The HTTP server provides a torrent index list with magnet support on the `/` route by default.



#### Features

- [x] HTTP Tracker
- [x] Compact peer format
- [x] Scrape
- [ ] UDP Tracker
- [ ] IPv6

#### Environment variables

- `ADDRESS` (0.0.0.0:9999)
- `ANNOUNCE_URL` (http://localhost:9999/announce)
  - Used for magnet links in the index vew
- `DSN` (postgres://tracker:tracker@localhost:5432/tracker)
- `TEMPLATE_PATH` (../templates/)
- `STATIC_PATH` (../static/)