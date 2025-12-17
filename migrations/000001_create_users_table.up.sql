CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username text UNIQUE NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP(0) WITH TIME ZONE,
    activated BOOL NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
);
