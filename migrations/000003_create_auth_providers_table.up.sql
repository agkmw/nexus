CREATE TABLE IF NOT EXISTS auth_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider          text NOT NULL,
    provider_user_id  text NOT NULL,
    email_at_provider citext,

    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),

    version integer NOT NULL DEFAULT 1

    UNIQUE(provider, provider_user_id),
    UNIQUE(provider, user_id),
);
