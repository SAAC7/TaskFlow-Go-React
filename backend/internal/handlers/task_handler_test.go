package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"taskflow-backend/internal/handlers"
	"taskflow-backend/internal/models"
	"taskflow-backend/internal/services"
)

// fakeRepo is an in-memory services.TaskRepository: the whole stack
// (handler -> service) is tested without MongoDB.
type fakeRepo struct {
	mu    sync.Mutex
	tasks map[bson.ObjectID]models.Task
}

func newFakeRepo() *fakeRepo { return &fakeRepo{tasks: map[bson.ObjectID]models.Task{}} }

func (f *fakeRepo) Create(_ context.Context, t *models.Task) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	t.ID = bson.NewObjectID()
	f.tasks[t.ID] = *t
	return nil
}

func (f *fakeRepo) FindAll(_ context.Context, _ models.ListFilter) ([]models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []models.Task{}
	for _, t := range f.tasks {
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeRepo) FindByID(_ context.Context, id bson.ObjectID) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return &t, nil
}

func (f *fakeRepo) Update(_ context.Context, id bson.ObjectID, p models.TaskPatch, at time.Time) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Completed != nil {
		t.Completed = *p.Completed
	}
	t.UpdatedAt = at
	f.tasks[id] = t
	return &t, nil
}

func (f *fakeRepo) Delete(_ context.Context, id bson.ObjectID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.tasks[id]; !ok {
		return models.ErrNotFound
	}
	delete(f.tasks, id)
	return nil
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewTaskService(newFakeRepo())
	handlers.NewTaskHandler(svc, logger).Register(mux)
	return mux
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCRUDFlow(t *testing.T) {
	h := newServer()

	rec := do(h, http.MethodPost, "/tasks", `{"title":"  Write tests  ","description":"cover everything"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d (%s)", rec.Code, rec.Body)
	}
	var created models.Task
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID.IsZero() || created.Title != "Write tests" || created.CreatedAt.IsZero() {
		t.Fatalf("create: unexpected task %+v", created)
	}
	id := created.ID.Hex()
	if loc := rec.Header().Get("Location"); loc != "/tasks/"+id {
		t.Fatalf("bad Location header %q", loc)
	}

	if rec := do(h, http.MethodGet, "/tasks/"+id, ""); rec.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d", rec.Code)
	}

	rec = do(h, http.MethodPatch, "/tasks/"+id, `{"completed":true}`)
	var patched models.Task
	_ = json.NewDecoder(rec.Body).Decode(&patched)
	if rec.Code != http.StatusOK || !patched.Completed || patched.Title != "Write tests" {
		t.Fatalf("patch: unexpected result %d %+v", rec.Code, patched)
	}

	if rec := do(h, http.MethodPut, "/tasks/"+id, `{"title":"New title","completed":false}`); rec.Code != http.StatusOK {
		t.Fatalf("put: want 200, got %d", rec.Code)
	}

	if rec := do(h, http.MethodDelete, "/tasks/"+id, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}
	if rec := do(h, http.MethodGet, "/tasks/"+id, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: want 404, got %d", rec.Code)
	}
}

func TestErrorResponses(t *testing.T) {
	h := newServer()
	missing := bson.NewObjectID().Hex() // valid format, but does not exist

	tests := []struct {
		name         string
		method, path string
		body         string
		want         int
	}{
		{"blank title", http.MethodPost, "/tasks", `{"title":"   "}`, http.StatusUnprocessableEntity},
		{"malformed json", http.MethodPost, "/tasks", `{"title":`, http.StatusBadRequest},
		{"empty body", http.MethodPost, "/tasks", ``, http.StatusBadRequest},
		{"wrong type", http.MethodPost, "/tasks", `{"title":123}`, http.StatusBadRequest},
		{"empty patch", http.MethodPatch, "/tasks/" + missing, `{}`, http.StatusUnprocessableEntity},
		{"invalid id", http.MethodGet, "/tasks/not-an-id", ``, http.StatusBadRequest},
		{"not found", http.MethodDelete, "/tasks/" + missing, ``, http.StatusNotFound},
		{"bad completed filter", http.MethodGet, "/tasks?completed=maybe", ``, http.StatusBadRequest},
		{"limit too large", http.MethodGet, "/tasks?limit=1000", ``, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(h, tc.method, tc.path, tc.body); rec.Code != tc.want {
				t.Fatalf("want %d, got %d (%s)", tc.want, rec.Code, rec.Body)
			}
		})
	}
}

// A client must not be able to choose the id or the timestamps.
func TestClientCannotForgeID(t *testing.T) {
	h := newServer()
	forged := bson.NewObjectID().Hex()

	rec := do(h, http.MethodPost, "/tasks", `{"id":"`+forged+`","title":"x"}`)
	var created models.Task
	_ = json.NewDecoder(rec.Body).Decode(&created)
	if created.ID.Hex() == forged {
		t.Fatal("server accepted a client-provided id")
	}
}
