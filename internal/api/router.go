package api

import (
	"log/slog"
	"net/http"

	"github.com/ibxbit/ledger/internal/store"
)

// NewServer wires routes and middleware into one http.Handler.
// Order matters: Recover is outermost so it catches panics from
// everything inside, including the logger.
func NewServer(logger *slog.Logger, st *store.Store) http.Handler {
	s := &Server{logger: logger, store: st}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /accounts", s.handleCreateAccount)
	mux.HandleFunc("GET /accounts/{id}", s.handleGetAccount)
	mux.HandleFunc("GET /boom", s.handleBoom) // TEMPORARY panic tester

	var handler http.Handler = mux
	handler = Logging(logger, handler)
	handler = Recover(logger, handler)
	return handler
}
