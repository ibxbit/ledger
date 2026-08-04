# ledger

A double-entry wallet/payments API. Go + PostgreSQL, stdlib HTTP, raw SQL — no frameworks, no ORM.

Built as part of COMEBACK_2026: production patterns, done manually, understood deeply.

## Stack

- Go 1.24 (`net/http` stdlib)
- PostgreSQL 17 (Docker), accessed via `psql` + `pgx/v5`
- Raw SQL migrations

## Run

```sh
# start the database
docker compose up -d

# psql shell into it
docker exec -it ledger-db psql -U ledger

# run the API (once it exists)
go run ./cmd/api
```

## Design

- **Double-entry:** every transfer writes two `entries` rows (−amount sender, +amount receiver). An account's balance is `SUM(entries.amount)` — there is no balance column to drift out of sync.
- **Money is BIGINT minor units** (cents). Floats never touch money.
- See `DECISIONS.md` for the reasoning behind each choice.
