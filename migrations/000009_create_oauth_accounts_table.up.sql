CREATE TABLE oauth_accounts(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider IN ('google', 'github', 'apple')),
    provider_user_id text NOT NULL,
    email_snapshot citext,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_account ON oauth_accounts(account_id);

