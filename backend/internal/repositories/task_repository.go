package repositories

import (
	"context"

	"taskflow-backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TaskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) *TaskRepository {
	return &TaskRepository{
		collection: db.Collection("tasks"),
	}
}

func (r *TaskRepository) Create(
	ctx context.Context,
	task *models.Task,
) error {

	result, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(bson.ObjectID)

	return nil
}

func (r *TaskRepository) FindAll(
	ctx context.Context,
) ([]models.Task, error) {

	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var tasks []models.Task

	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

func (r *TaskRepository) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (*models.Task, error) {

	var task models.Task

	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
	).Decode(&task)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	task *models.Task,
) error {

	filter := bson.D{
		{Key: "_id", Value: id},
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "title", Value: task.Title},
				{Key: "description", Value: task.Description},
				{Key: "completed", Value: task.Completed},
			},
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *TaskRepository) Delete(
	ctx context.Context,
	id bson.ObjectID,
) error {

	result, err := r.collection.DeleteOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
	)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
