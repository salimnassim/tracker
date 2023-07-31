# tracker
BitTorrent tracker which relies on PostgreSQL for persistence. All announced torrents will be tracked automatically.
The HTTP server provides a torrent index list with magnet support on the `/` route by default.

#### Features

- [x] HTTP Tracker
- [x] Compact peer format
- [ ] Scrape
- [ ] UDP Tracker
- [ ] Upload metadata files
- [ ] IPv6
- [ ] GeoIP matching