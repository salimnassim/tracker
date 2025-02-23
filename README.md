# tracker

This is a BitTorrent tracker that provides both an HTTP API index list and a UDP tracker for torrent clients.

## Configuration

The tracker behavior is configured using environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `BT_HTTP_ADDRESS` | The address the HTTP server binds to | `0.0.0.0` |
| `BT_HTTP_PORT` | The port for the HTTP API | `8080` |
| `BT_UDP_ADDRESS` | The address the UDP tracker binds to | `0.0.0.0` |
| `BT_UDP_PORT` | The port for the UDP tracker | `6118` |
| `BT_UDP_URL` | The full UDP tracker URL (e.g., `udp://tracker.example.com:6118/announce`) | `udp://localhost:6118/announce` |
