-- name: RegisterTorrent :one
INSERT INTO torrents (info_hash, magnet, completed)
VALUES (?, ?, ?)
ON CONFLICT (info_hash) DO UPDATE SET info_hash = excluded.info_hash
RETURNING *;

-- name: GetTorrent :one
SELECT * FROM torrents
WHERE info_hash = ?
LIMIT 1;

-- name: ListTorrents :many
SELECT * FROM torrents
ORDER BY info_hash;

-- name: IncrementCompleted :exec
UPDATE torrents
SET completed = completed + 1
WHERE info_hash = ?;
