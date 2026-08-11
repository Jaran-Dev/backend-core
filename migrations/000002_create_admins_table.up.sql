CREATE TABLE admins(
    account_id uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);

