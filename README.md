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

#### Resources

https://wiki.theory.org/BitTorrentSpecification#Tracker_HTTP.2FHTTPS_Protocol
https://www.bittorrent.org/beps/bep_0015.html
https://www.rasterbar.com/products/libtorrent/udp_tracker_protocol.html#connecting