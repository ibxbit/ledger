package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// Account mirrors the accounts table. Capitalized fields are exported;
// json tags control the wire format.
type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"` // minor units (cents), per DECISIONS.md
	CreatedAt time.Time `json:"created_at"`
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

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /accounts — fake store until the DB lands on Day 5.
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
		ID:        1,
		UserID:    req.UserID,
		Currency:  req.Currency,
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}
	writeJSON(w, http.StatusCreated, acc)
}

// GET /accounts/{id}
func handleGetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "id must be a positive integer"})
		return
	}
	if id != 1 { // fake store: only account 1 exists today
		writeJSON(w, http.StatusNotFound, apiError{Error: "account not found"})
		return
	}
	writeJSON(w, http.StatusOK, Account{
		ID: 1, UserID: 1, Currency: "USD", Balance: -5000,
		CreatedAt: time.Now().UTC(),
	})
}

// GET /boom — deliberately panics so we can watch Recover do its job.
// TEMPORARY: delete before this ever ships anywhere real.
func handleBoom(w http.ResponseWriter, r *http.Request) {
	panic("boom: deliberate test panic")
}
