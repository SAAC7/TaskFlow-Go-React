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

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler { return &HealthHandler{db: db} }

func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Live)
	mux.HandleFunc("GET /ready", h.Ready)
}

// GET /health: the process is alive (Render's health check uses this one).
func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /ready: the process is alive AND MongoDB answers.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
