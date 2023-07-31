CREATE TABLE IF NOT EXISTS public.torrents
(
    id uuid NOT NULL,
    info_hash bytea NOT NULL,
    completed integer NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT torrents_pkey PRIMARY KEY (id),
    CONSTRAINT torrents_info_hash_key UNIQUE (info_hash)
);

CREATE TABLE IF NOT EXISTS public.peers
(
    id uuid NOT NULL,
    torrent_id uuid NOT NULL,
    peer_id bytea NOT NULL,
    ip inet NOT NULL,
    port integer NOT NULL,
    uploaded integer NOT NULL DEFAULT 0,
    downloaded integer NOT NULL DEFAULT 0,
    "left" integer NOT NULL,
    "event" text COLLATE pg_catalog."default" NOT NULL,
    "key" text COLLATE pg_catalog."default",
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT peers_pkey PRIMARY KEY (id),
    CONSTRAINT peers_torrent_id_peer_id_key UNIQUE (torrent_id, peer_id),
    CONSTRAINT peers_torrent_id_fkey FOREIGN KEY (torrent_id)
        REFERENCES public.torrents (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

ALTER TABLE IF EXISTS public.peers
    OWNER to tracker;

ALTER TABLE IF EXISTS public.torrents
    OWNER to tracker;