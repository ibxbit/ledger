package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors: handlers translate these to HTTP statuses without
// knowing anything about SQL.
var (
	ErrNotFound     = errors.New("not found")
	ErrUserNotFound = errors.New("user not found")
)

type Account struct {
	ID        int64
	UserID    int64
	Currency  string
	Balance   int64 // minor units, computed from entries — never stored
	CreatedAt time.Time
}

// Store owns the connection pool. All queries go through methods on it,
// so SQL lives in exactly one package.
type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	// pgxpool connects lazily; Ping forces a real connection now so a bad
	// DB URL fails at startup, not on the first request.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

// CreateAccount inserts a row and returns it. $1/$2 placeholders mean the
// driver sends values separately from SQL text — SQL injection is
// structurally impossible.
func (s *Store) CreateAccount(ctx context.Context, userID int64, currency string) (Account, error) {
	var a Account
	err := s.pool.QueryRow(ctx,
		`INSERT INTO accounts (user_id, currency)
		 VALUES ($1, $2)
		 RETURNING id, user_id, currency, created_at`,
		userID, currency,
	).Scan(&a.ID, &a.UserID, &a.Currency, &a.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23503 = foreign_key_violation: the user_id points at no real user.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return Account{}, ErrUserNotFound
		}
		return Account{}, err
	}
	return a, nil
}

// GetAccount returns the account with its balance computed live from
// entries. LEFT JOIN + COALESCE so an account with no entries yet shows
// balance 0 instead of disappearing or returning NULL.
func (s *Store) GetAccount(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := s.pool.QueryRow(ctx,
		`SELECT a.id, a.user_id, a.currency, a.created_at,
		        COALESCE(SUM(e.amount), 0) AS balance
		 FROM accounts a
		 LEFT JOIN entries e ON e.account_id = a.id
		 WHERE a.id = $1
		 GROUP BY a.id`,
		id,
	).Scan(&a.ID, &a.UserID, &a.Currency, &a.CreatedAt, &a.Balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}
