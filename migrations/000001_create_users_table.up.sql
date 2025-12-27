CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    username     text UNIQUE NOT NULL,
    display_name text        NOT NULL,
    -- TODO: Add Avatar URL

    email           citext UNIQUE,
    password_hash   bytea,         
    email_verified  bool,          

    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_login timestamp(0) with time zone,

    version integer NOT NULL DEFAULT 1
);
