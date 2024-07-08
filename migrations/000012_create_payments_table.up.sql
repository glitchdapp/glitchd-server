CREATE TABLE IF NOT EXISTS payments (
    id UUID NOT NULL,
    session_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    entity_type TEXT,
    status TEXT NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);