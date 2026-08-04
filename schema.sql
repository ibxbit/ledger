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

-- transfers: the intent of a money movement
CREATE TABLE transfers (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    from_account BIGINT NOT NULL REFERENCES accounts(id),
    to_account   BIGINT NOT NULL REFERENCES accounts(id),
    amount       BIGINT NOT NULL CHECK (amount > 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_account <> to_account)
);

-- entries: the double-entry truth; balance = SUM(amount) per account.
-- No positivity check: the sender's entry is negative by design.
CREATE TABLE entries (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id  BIGINT NOT NULL REFERENCES accounts(id),
    transfer_id BIGINT NOT NULL REFERENCES transfers(id),
    amount      BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
