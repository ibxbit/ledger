package api

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code —
// the raw interface gives handlers no way to read back what was written.
// Embedding the interface means we inherit all its methods for free and
// override only WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging wraps any handler and logs one structured line per request.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, req)

		logger.Info("request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// Recover turns a handler panic into a 500 response instead of a dead server.
// defer runs even during a panic; recover() stops the panic from propagating.
func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "err", v, "path", req.URL.Path)
				writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
			}
		}()
		next.ServeHTTP(w, req)
	})
}
