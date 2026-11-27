package handlers

import (
	"context"
	"net/http"
	"time"
)

// Pinger is anything that can check its connectivity (the repository does).
type Pinger interface {
	Ping(ctx context.Context) error
}
// HealthHandler serves the liveness and readiness probes of the API.
type HealthHandler struct {
	db Pinger
}

// NewHealthHandler builds a HealthHandler that uses db to check the database connection.
func NewHealthHandler(db Pinger) *HealthHandler { return &HealthHandler{db: db} }

// Register mounts GET /health and GET /ready on the given mux.
func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Live)
	mux.HandleFunc("GET /ready", h.Ready)
}

// GET /health: the process is alive (Render's health check uses this one).
func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready handles GET /ready: it answers 200 only if the process is alive AND
// MongoDB answers a ping within 2 seconds (503 otherwise).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
