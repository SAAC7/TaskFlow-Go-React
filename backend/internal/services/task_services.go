// Package services holds the business logic: validation, timestamps and the
// rules of the application. It does not know about HTTP or about MongoDB.
package services

import (
	"context"
	"time"

	"taskflow-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TaskRepository is what the service NEEDS from the storage layer.
//
// In Go the interface is declared by the code that USES it (the consumer), not
// by the one that implements it. Any type with these methods satisfies it, so
// the MongoDB repository works in production and a tiny fake works in tests.
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindAll(ctx context.Context, filter models.ListFilter) ([]models.Task, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*models.Task, error)
	Update(ctx context.Context, id bson.ObjectID, patch models.TaskPatch, updatedAt time.Time) (*models.Task, error)
	Delete(ctx context.Context, id bson.ObjectID) error
}

// TaskService contains the business rules of the tasks and uses a
// TaskRepository to store them.
type TaskService struct {
	repository TaskRepository
}

// NewTaskService builds a TaskService that stores its data in the given repository.
func NewTaskService(repository TaskRepository) *TaskService {
	return &TaskService{repository: repository}
}

// now returns the current UTC time truncated to milliseconds (MongoDB's precision).
func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

// Create validates the input, sets the creation and update timestamps and
// stores the new task. It returns a *models.ValidationError if the input is invalid.
func (s *TaskService) Create(ctx context.Context, in models.TaskInput) (*models.Task, error) {
	if problems := in.Validate(); len(problems) > 0 {
		return nil, &models.ValidationError{Fields: problems}
	}

	t := now()
	task := &models.Task{
		Title:       in.Title,
		Description: in.Description,
		Completed:   in.Completed,
		CreatedAt:   t,
		UpdatedAt:   t,
	}
	if err := s.repository.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// FindAll returns the tasks that match the filter (newest first).
func (s *TaskService) FindAll(ctx context.Context, filter models.ListFilter) ([]models.Task, error) {
	return s.repository.FindAll(ctx, filter)
}

// FindByID converts the id received from the URL into an ObjectID and returns
// that task (models.ErrInvalidID or models.ErrNotFound on failure).
func (s *TaskService) FindByID(ctx context.Context, id string) (*models.Task, error) {
	oid, err := models.ParseID(id)
	if err != nil {
		return nil, err
	}
	return s.repository.FindByID(ctx, oid)
}

// Replace implements PUT: title is required and every field is overwritten.
func (s *TaskService) Replace(ctx context.Context, id string, in models.TaskInput) (*models.Task, error) {
	oid, err := models.ParseID(id)
	if err != nil {
		return nil, err
	}
	if problems := in.Validate(); len(problems) > 0 {
		return nil, &models.ValidationError{Fields: problems}
	}
	return s.repository.Update(ctx, oid, in.ToPatch(), now())
}

// Patch implements PATCH: only the fields that were sent are changed.
func (s *TaskService) Patch(ctx context.Context, id string, patch models.TaskPatch) (*models.Task, error) {
	oid, err := models.ParseID(id)
	if err != nil {
		return nil, err
	}
	if problems := patch.Validate(); len(problems) > 0 {
		return nil, &models.ValidationError{Fields: problems}
	}
	return s.repository.Update(ctx, oid, patch, now())
}

// Delete converts the id received from the URL into an ObjectID and removes
// that task (models.ErrInvalidID or models.ErrNotFound on failure).
func (s *TaskService) Delete(ctx context.Context, id string) error {
	oid, err := models.ParseID(id)
	if err != nil {
		return err
	}
	return s.repository.Delete(ctx, oid)
}
