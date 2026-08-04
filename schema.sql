-- ledger schema
-- Authored by hand in psql; this file is the canonical record.
-- Order matters: referenced tables first.

CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    currency   TEXT NOT NULL CHECK (char_length(currency) = 3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- transfers: the intent of a money movement (user writes this table)
-- entries: the double-entry truth; balance = SUM(amount) per account (user writes this table)
