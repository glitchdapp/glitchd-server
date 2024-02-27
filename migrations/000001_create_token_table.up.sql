CREATE TABLE IF NOT EXISTS tokens (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);
