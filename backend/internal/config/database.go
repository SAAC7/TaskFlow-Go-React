package config

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ConnectDB connects to MongoDB Atlas and verifies the connection.
// It returns the client too, so main can disconnect it on shutdown.
func ConnectDB(ctx context.Context, uri, databaseName string) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		return nil, nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("ping mongodb (check URI, credentials and Atlas Network Access): %w", err)
	}

	return client, client.Database(databaseName), nil
}
