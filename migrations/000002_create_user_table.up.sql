CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    biography TEXT NOT NULL,
    stripe_customer_id TEXT NOT NULL,
    photo TEXT NOT NULL,
    cover TEXT,
    description TEXT NOT NULL,
    links TEXT NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
