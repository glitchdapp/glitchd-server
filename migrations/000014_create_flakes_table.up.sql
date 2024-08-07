CREATE TABLE IF NOT EXISTS flakes (
    id UUID NOT NULL,
    user_id TEXT NOT NULL UNIQUE,
    amount INTEGER NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channel_flakes (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);
