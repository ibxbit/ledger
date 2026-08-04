package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ibxbit/ledger/internal/api"
)

// main is wiring only: build dependencies, hand them to the api package, run.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	handler := api.NewServer(logger)

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
