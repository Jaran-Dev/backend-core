CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    email CITEXT UNIQUE NOT NULL,
    password_hash text,
    role TEXT NOT NULL CHECK (ROLE IN ('user', 'admin')),
    status text NOT NULL DEFAULT 'pending_verification' CHECK (status IN ('active', 'banned', 'locked', 'pending_verification')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_email ON accounts(email);

