CREATE TABLE IF NOT EXISTS payments (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    order_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);
