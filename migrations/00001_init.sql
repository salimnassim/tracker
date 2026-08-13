-- +goose Up
CREATE TABLE torrents (
  info_hash TEXT PRIMARY KEY,
  magnet    TEXT NOT NULL,
  completed INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE peers (
  info_hash  TEXT NOT NULL REFERENCES torrents(info_hash) ON DELETE CASCADE,
  peer_id    TEXT NOT NULL,
  downloaded INTEGER NOT NULL DEFAULT 0,
  "left"     INTEGER NOT NULL DEFAULT 0,
  uploaded   INTEGER NOT NULL DEFAULT 0,
  event      INTEGER NOT NULL DEFAULT 0,
  ip         INTEGER NOT NULL DEFAULT 0,
  port       INTEGER NOT NULL DEFAULT 0,
  "key"      INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (info_hash, peer_id)
);

CREATE INDEX idx_peers_updated_at ON peers(updated_at);

-- +goose Down
DROP TABLE peers;
DROP TABLE torrents;
