package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Account mirrors the accounts table. The `json:"..."` tags control the field
// names when this struct is encoded to JSON (Go fields are CapitalCase because
// only capitalized identifiers are exported/visible outside the package).
type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"` // minor units (cents), per DECISIONS.md
	CreatedAt time.Time `json:"created_at"`
}

// createAccountRequest is the shape of the POST /accounts body.
type createAccountRequest struct {
	UserID   int64  `json:"user_id"`
	Currency string `json:"currency"`
}

// apiError is the one error shape every failure returns.
type apiError struct {
	Error string `json:"error"`
}

// writeJSON is the single place responses get encoded: set the header, set the
// status code, encode the value. Every handler uses it — one behavior, one place.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleCreateAccount: POST /accounts
// Today it fakes the DB: validates input and echoes back what would be created.
func handleCreateAccount(w http.ResponseWriter, r *http.Request) {
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

	acc := Account{
		ID:        1, // fake — Postgres will assign real ids on Day 5
		UserID:    req.UserID,
		Currency:  req.Currency,
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}
	writeJSON(w, http.StatusCreated, acc)
}

// handleGetAccount: GET /accounts/{id}
func handleGetAccount(w http.ResponseWriter, r *http.Request) {
	// r.PathValue reads the {id} segment matched by the route pattern.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "id must be a positive integer"})
		return
	}
	if id != 1 { // fake store: only account 1 "exists" today
		writeJSON(w, http.StatusNotFound, apiError{Error: "account not found"})
		return
	}
	writeJSON(w, http.StatusOK, Account{
		ID: 1, UserID: 1, Currency: "USD", Balance: -5000,
		CreatedAt: time.Now().UTC(),
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	// Go 1.22+ patterns: "METHOD /path". {id} is a wildcard segment.
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /accounts", handleCreateAccount)
	mux.HandleFunc("GET /accounts/{id}", handleGetAccount)

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
