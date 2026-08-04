package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ibxbit/ledger/internal/store"
)

// Server holds the handlers' dependencies. Handlers are methods on it,
// so they reach the store and logger without globals.
type Server struct {
	logger *slog.Logger
	store  *store.Store
}

// Account is the wire format (JSON). Kept separate from store.Account so
// the API shape can evolve without touching the DB layer.
type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"` // minor units (cents), per DECISIONS.md
	CreatedAt time.Time `json:"created_at"`
}

func toAPIAccount(a store.Account) Account {
	return Account{
		ID:        a.ID,
		UserID:    a.UserID,
		Currency:  a.Currency,
		Balance:   a.Balance,
		CreatedAt: a.CreatedAt,
	}
}

type createAccountRequest struct {
	UserID   int64  `json:"user_id"`
	Currency string `json:"currency"`
}

// apiError is the one error shape every failure returns.
type apiError struct {
	Error string `json:"error"`
}

// writeJSON is the single place responses get encoded.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /accounts — inserts a real row now.
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid JSON body"})
		return
	}
	if req.UserID <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "user_id is required"})
		return
	}
	if len(req.Currency) != 3 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "currency must be a 3-letter code"})
		return
	}

	// r.Context() flows into the query: if the client disconnects, the
	// DB work gets cancelled too.
	acc, err := s.store.CreateAccount(r.Context(), req.UserID, req.Currency)
	if errors.Is(err, store.ErrUserNotFound) {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "user does not exist"})
		return
	}
	if err != nil {
		s.logger.Error("create account", "err", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusCreated, toAPIAccount(acc))
}

// GET /accounts/{id} — balance computed live from entries.
func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "id must be a positive integer"})
		return
	}

	acc, err := s.store.GetAccount(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, apiError{Error: "account not found"})
		return
	}
	if err != nil {
		s.logger.Error("get account", "err", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, toAPIAccount(acc))
}

// GET /boom — deliberately panics so we can watch Recover do its job.
// TEMPORARY: delete before this ever ships anywhere real.
func (s *Server) handleBoom(w http.ResponseWriter, r *http.Request) {
	panic("boom: deliberate test panic")
}
