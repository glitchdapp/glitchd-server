CREATE TABLE IF NOT EXISTS logs (
    id UUID NOT NULL,
    data TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);
