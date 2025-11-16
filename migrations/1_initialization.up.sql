/* Наверное, что-то такое */
CREATE TABLE IF NOT EXISTS public.metrics (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('gauge', 'counter')),
    delta BIGINT,
    value DOUBLE PRECISION,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
