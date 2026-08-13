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
| `BT_UDP_URL` | The full UDP tracker URL | `udp://localhost:6118/announce` |
| `PEER_INTERVAL` | Peer expiry check interval | `1200` (seconds) |
| `PEER_LIFETIME` | Peer expiry lifetime | `3600` (seconds) |
| `TRACKER_DB_PATH` | Path to the SQLite database file | `/data/tracker.db` |

## HTTP Server

The HTTP server provides a list of torrents and their peers.

```
{
   "ae4a048e19e6ed8ce31bc5a37f0ca03a533998a4":{
      "magnet":"magnet:?xt=urn:btih:ae4a048e19e6ed8ce31bc5a37f0ca03a533998a4\u0026tr=udp://127.0.0.1:6881",
      "completed":0,
      "peers":{
         "-TR3000-hi33owz1ajb7":{
            "downloaded":0,
            "left":0,
            "uploaded":0,
            "event":0
         },
         "-qB5040-1LwzYpZa67!D":{
            "downloaded":0,
            "left":0,
            "uploaded":0,
            "event":2
         }
      }
   },
   "ef278c16ebe1d63f5d1ea4d271a9a06d23a1f10b":{
      "magnet":"magnet:?xt=urn:btih:ef278c16ebe1d63f5d1ea4d271a9a06d23a1f10b\u0026tr=udp://127.0.0.1:6881",
      "completed":5,
      "peers":{
         "-qB5040-UIkh)Z270fKP":{
            "downloaded":0,
            "left":14128,
            "uploaded":0,
            "event":2
         }
      }
   }
}
```