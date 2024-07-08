CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL,
    name TEXT,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    biography TEXT,
    stripe_customer_id TEXT UNIQUE,
    stripe_connected_link BOOLEAN DEFAULT FALSE,
    photo TEXT,
    cover TEXT,
    description TEXT,
    links TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
