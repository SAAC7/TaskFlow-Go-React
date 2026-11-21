// Package repositories is the ONLY place that talks to MongoDB.
package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"taskflow-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type TaskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) *TaskRepository {
	return &TaskRepository{collection: db.Collection("tasks")}
}

// EnsureIndexes creates the index that supports "newest first" listings.
// Creating an index that already exists is a no-op, so it is safe on every start.
func (r *TaskRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdAt", Value: -1}},
	})
	if err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}
	return nil
}

// Ping checks that MongoDB is reachable (used by GET /ready).
func (r *TaskRepository) Ping(ctx context.Context) error {
	return r.collection.Database().Client().Ping(ctx, readpref.Primary())
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	result, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	// The driver generated the _id; we copy it back so the caller has it.
	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return fmt.Errorf("unexpected inserted id type %T", result.InsertedID)
	}
	task.ID = id
	return nil
}

func (r *TaskRepository) FindAll(ctx context.Context, f models.ListFilter) ([]models.Task, error) {
	query := bson.D{}
	if f.Completed != nil {
		query = append(query, bson.E{Key: "completed", Value: *f.Completed})
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}). // newest first
		SetLimit(f.Limit).
		SetSkip(f.Skip)

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("decode tasks: %w", err)
	}
	if tasks == nil {
		tasks = []models.Task{} // JSON "[]" instead of "null"
	}
	return tasks, nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.Task, error) {
	var task models.Task

	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&task)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, models.ErrNotFound // translate the driver error to the domain error
	}
	if err != nil {
		return nil, fmt.Errorf("find task: %w", err)
	}
	return &task, nil
}

// Update applies the non-nil fields of the patch and returns the updated task
// in a single round trip (FindOneAndUpdate + ReturnDocument After).
func (r *TaskRepository) Update(ctx context.Context, id bson.ObjectID, p models.TaskPatch, updatedAt time.Time) (*models.Task, error) {
	set := bson.D{{Key: "updatedAt", Value: updatedAt}}
	if p.Title != nil {
		set = append(set, bson.E{Key: "title", Value: *p.Title})
	}
	if p.Description != nil {
		set = append(set, bson.E{Key: "description", Value: *p.Description})
	}
	if p.Completed != nil {
		set = append(set, bson.E{Key: "completed", Value: *p.Completed})
	}

	var task models.Task
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: set}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&task)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if result.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}
