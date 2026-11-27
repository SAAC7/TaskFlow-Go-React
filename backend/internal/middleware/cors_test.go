package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"taskflow-backend/internal/middleware"
)

// corsHandler wraps a dummy handler (always 200) with the CORS middleware and
// an allow-list of two origins, one of them written with a trailing slash.
func corsHandler() http.Handler {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return middleware.CORS([]string{"http://localhost:5173", "https://taskflow.vercel.app/"})(ok)
}

// TestCORSPreflightAllowed checks that a preflight OPTIONS request from an
// allowed origin gets 204 and the Access-Control-Allow-* headers.
func TestCORSPreflightAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
	req.Header.Set("Origin", "https://taskflow.vercel.app") // allow-list entry has a trailing slash
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	corsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://taskflow.vercel.app" {
		t.Fatalf("unexpected Allow-Origin %q", got)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("missing Allow-Methods")
	}
}


// TestCORSPreflightDenied checks that a preflight request from an origin that
// is not in the allow-list gets 403 and no CORS headers.
func TestCORSPreflightDenied(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	corsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("denied origin must not receive CORS headers")
	}
}

// TestCORSSimpleRequest checks that a normal (non-preflight) request from an
// allowed origin receives the Access-Control-Allow-Origin header.
func TestCORSSimpleRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	corsHandler().ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("unexpected Allow-Origin %q", got)
	}
}
