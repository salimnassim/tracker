ALTER TABLE public.peers ALTER COLUMN uploaded TYPE bigint;
ALTER TABLE public.peers ALTER COLUMN downloaded TYPE bigint;
ALTER TABLE public.peers ALTER COLUMN "left" TYPE bigint;