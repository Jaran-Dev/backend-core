CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source TEXT NOT NULL CHECK (source IN ('manual', 'gateway', 'promo')), -- Defense in depth
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at TIMESTAMPTZ, -- NULL means lifetime access
    provider_ref TEXT,    -- Transaction ID from gateway (empty if manual)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_active ON subscriptions(user_id) 
WHERE starts_at <= now() AND (ends_at IS NULL OR ends_at > now());