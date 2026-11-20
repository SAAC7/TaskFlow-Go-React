package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"taskflow-backend/internal/models"
	"taskflow-backend/internal/services"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(
	service *services.TaskService,
) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func respondJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err := h.service.Create(r.Context(), &task)

	if errors.Is(err, services.ErrInvalidTask) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not create task")
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.FindAll(r.Context())

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not retrieve tasks")
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}
func (h *TaskHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.service.FindByID(r.Context(), id)

	if errors.Is(err, mongo.ErrNoDocuments) {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not retrieve task")
		return
	}

	respondJSON(w, http.StatusOK, task)
}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = h.service.Update(r.Context(), id, &task)

	if errors.Is(err, services.ErrInvalidTask) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not update task")
		return
	}

	task.ID = id

	respondJSON(w, http.StatusOK, task)
}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	err = h.service.Delete(r.Context(), id)

	if errors.Is(err, mongo.ErrNoDocuments) {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not delete task")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Task deleted successfully",
	})
}
