CREATE TABLE IF NOT EXISTS public.announce_log
(
    id uuid NOT NULL,
    info_hash bytea NOT NULL,
    peer_id bytea NOT NULL,
    event text COLLATE pg_catalog."default" NOT NULL,
    ip inet NOT NULL,
    port integer NOT NULL,
    key text COLLATE pg_catalog."default" NOT NULL,
    uploaded bigint NOT NULL,
    downloaded bigint NOT NULL,
    "left" bigint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT announce_log_pkey PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.announce_log
    OWNER to tracker;