CREATE TABLE users(
    account_id uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    display_name text NOT NULL,
    username CITEXT UNIQUE,
    bio text,
    headline text,
    pronouns varchar(20),
    location varchar(100),
    date_of_birth date,
    avatar_url text,
    avatar_status varchar(20) NOT NULL DEFAULT 'pending',
    cover_url text,
    cover_status varchar(20) NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version int NOT NULL DEFAULT 1
);

CREATE INDEX idx_users_username ON users(username);

