package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"taskflow-backend/internal/middleware"
	"taskflow-backend/internal/models"
	"taskflow-backend/internal/services"
)

const (
	requestTimeout = 5 * time.Second // max time for one request's database work
	defaultLimit   = 50
	maxLimit       = 100
)

// TaskHandler is the HTTP layer of the tasks: it translates requests into
// service calls and service results into JSON responses.
type TaskHandler struct {
	service *services.TaskService
	log     *slog.Logger
}

// NewTaskHandler builds a TaskHandler with the service and the logger it depends on.
func NewTaskHandler(service *services.TaskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{service: service, log: log}
}

// Register mounts the task routes (Go 1.22 "METHOD /path/{param}" patterns).
func (h *TaskHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("GET /tasks", h.FindAll)
	mux.HandleFunc("GET /tasks/{id}", h.FindByID)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)
	mux.HandleFunc("PATCH /tasks/{id}", h.Patch)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)
}

// Create handles POST /tasks: it reads the JSON body, creates the task and
// answers 201 Created with the new task and a Location header.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.TaskInput
	if err := decodeJSON(w, r, &in); err != nil {
		respondDecodeError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	task, err := h.service.Create(ctx, in)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	w.Header().Set("Location", "/tasks/"+task.ID.Hex())
	respondJSON(w, http.StatusCreated, task)
}

// FindAll handles GET /tasks?completed=true&limit=20&page=2: it returns the
// list of tasks, newest first, applying the optional filter and pagination.
func (h *TaskHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r.URL.Query())
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	tasks, err := h.service.FindAll(ctx, filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, tasks)
}

// FindByID handles GET /tasks/{id}: it returns a single task (404 if it does not exist).
func (h *TaskHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	task, err := h.service.FindByID(ctx, r.PathValue("id"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

// PUT /tasks/{id}: replaces title, description and completed (title required).
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in models.TaskInput
	if err := decodeJSON(w, r, &in); err != nil {
		respondDecodeError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	task, err := h.service.Replace(ctx, r.PathValue("id"), in)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

// PATCH /tasks/{id}: changes only the fields that were sent
// (e.g. {"completed": true} to toggle a task).
func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	var patch models.TaskPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		respondDecodeError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	task, err := h.service.Patch(ctx, r.PathValue("id"), patch)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

// Delete handles DELETE /tasks/{id}: it removes the task and answers
// 204 No Content (no body).
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	if err := h.service.Delete(ctx, r.PathValue("id")); err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleError is the ONLY place that maps domain errors to HTTP status codes.
// Unknown errors are logged in full but the client only sees a generic message,
// so internal details (database messages) never leak.
func (h *TaskHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *models.ValidationError

	switch {
	case errors.As(err, &validationErr):
		respondError(w, http.StatusUnprocessableEntity, "validation_failed", "One or more fields are invalid", validationErr.Fields)
	case errors.Is(err, models.ErrInvalidID):
		respondError(w, http.StatusBadRequest, "invalid_id", "The task id is not valid", nil)
	case errors.Is(err, models.ErrNotFound):
		respondError(w, http.StatusNotFound, "not_found", "Task not found", nil)
	default:
		h.log.Error("internal error", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		respondError(w, http.StatusInternalServerError, "internal_error", "Something went wrong", nil)
	}
}

// parseListFilter reads and validates the query string of GET /tasks
// (completed, limit and page) and converts it into a models.ListFilter.
func parseListFilter(q url.Values) (models.ListFilter, error) {
	f := models.ListFilter{Limit: defaultLimit}

	if v := q.Get("completed"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return f, errors.New("completed must be true or false")
		}
		f.Completed = &b
	}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxLimit {
			return f, fmt.Errorf("limit must be a number between 1 and %d", maxLimit)
		}
		f.Limit = int64(n)
	}

	page := 1
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return f, errors.New("page must be a number greater than or equal to 1")
		}
		page = n
	}
	f.Skip = int64(page-1) * f.Limit

	return f, nil
}
