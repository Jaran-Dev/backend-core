CREATE TABLE subscriptions(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    source text NOT NULL CHECK (source IN ('manual', 'gateway', 'promo')),
    starts_at timestamptz NOT NULL DEFAULT now(),
    ends_at timestamptz,
    provider_ref text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_active ON subscriptions(account_id)
WHERE
    starts_at <= now() AND (ends_at IS NULL OR ends_at > now());

