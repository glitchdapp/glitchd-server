CREATE TABLE IF NOT EXISTS memberships (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    gifter TEXT,
    is_gift BOOLEAN NOT NULL DEFAULT false,
    tier TEXT,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS membership_details (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    tier INTEGER NOT NULL DEFAULT 1,
    name TEXT,
    description TEXT,
    badges TEXT[],
    cost INTEGER NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);

