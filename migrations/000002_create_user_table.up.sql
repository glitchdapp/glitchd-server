CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    stripe_customer_id TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW() 
);
