package services

import (
	"context"
	"errors"
	"strings"

	"taskflow-backend/internal/models"
	"taskflow-backend/internal/repositories"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidTask = errors.New("task title is required")

type TaskService struct {
	repository *repositories.TaskRepository
}

func NewTaskService(
	repository *repositories.TaskRepository,
) *TaskService {
	return &TaskService{
		repository: repository,
	}
}

func (s *TaskService) Create(
	ctx context.Context,
	task *models.Task,
) error {

	task.Title = strings.TrimSpace(task.Title)

	if task.Title == "" {
		return ErrInvalidTask
	}

	task.ID = bson.NilObjectID

	return s.repository.Create(ctx, task)
}

func (s *TaskService) FindAll(
	ctx context.Context,
) ([]models.Task, error) {

	return s.repository.FindAll(ctx)
}

func (s *TaskService) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (*models.Task, error) {

	return s.repository.FindByID(ctx, id)
}

func (s *TaskService) Update(
	ctx context.Context,
	id bson.ObjectID,
	task *models.Task,
) error {

	task.Title = strings.TrimSpace(task.Title)

	if task.Title == "" {
		return ErrInvalidTask
	}

	return s.repository.Update(ctx, id, task)
}

func (s *TaskService) Delete(
	ctx context.Context,
	id bson.ObjectID,
) error {

	return s.repository.Delete(ctx, id)
}
