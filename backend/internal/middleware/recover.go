package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover turns a panic inside a handler into a 500 JSON response instead of
// dropping the connection, and logs the stack trace for debugging.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
						"request_id", GetRequestID(r.Context()),
					)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":{"code":"internal_error","message":"Something went wrong"}}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
