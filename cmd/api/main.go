package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ibxbit/ledger/internal/api"
	"github.com/ibxbit/ledger/internal/store"
)

// main is wiring only: config, dependencies, run.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Config from env with a local-dev default. In real deploys the env
	// var is required; the default only exists for developer convenience.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ledger:secret@localhost:5432/ledger"
	}

	st, err := store.New(context.Background(), dbURL)
	if err != nil {
		logger.Error("connecting to database", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	handler := api.NewServer(logger, st)

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
