-- name: GetPeer :one
SELECT * FROM peers
WHERE info_hash = ? AND peer_id = ?
LIMIT 1;

-- name: UpsertPeer :exec
INSERT INTO peers (
  info_hash, peer_id, downloaded, "left", uploaded, event, ip, port, "key", updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (info_hash, peer_id) DO UPDATE SET
  downloaded = excluded.downloaded,
  "left" = excluded."left",
  uploaded = excluded.uploaded,
  event = excluded.event,
  ip = excluded.ip,
  port = excluded.port,
  "key" = excluded."key",
  updated_at = excluded.updated_at;

-- name: ListPeersByTorrent :many
SELECT * FROM peers
WHERE info_hash = ?;

-- name: ListAllPeers :many
SELECT * FROM peers;

-- name: DeleteExpiredPeers :execrows
DELETE FROM peers
WHERE updated_at < ?;
