package api

import (
	"log/slog"
	"net/http"
)

// NewServer wires routes and middleware into one http.Handler.
// Order matters: Recover is outermost so it catches panics from
// everything inside, including the logger.
func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /accounts", handleCreateAccount)
	mux.HandleFunc("GET /accounts/{id}", handleGetAccount)
	mux.HandleFunc("GET /boom", handleBoom) // TEMPORARY panic tester

	var handler http.Handler = mux
	handler = Logging(logger, handler)
	handler = Recover(logger, handler)
	return handler
}
