package app

import (
	"context"
	"time"

	"github.com/guttosm/user-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitMongo initializes a connection to MongoDB and returns the client,
// selected database instance, and a cleanup function.
//
// Parameters:
//   - cfg (*config.Config): The application configuration
//
// Returns:
//   - *mongo.Client: MongoDB client
//   - *mongo.Database: MongoDB database handle
//   - func(): Cleanup function to close client connection and cancel context
//   - error: Any connection or ping error
func InitMongo(cfg *config.Config) (*mongo.Client, *mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	db := client.Database(cfg.Mongo.Name)

	cleanup := func() {
		cancel()
		_ = client.Disconnect(context.Background())
	}

	return client, db, cleanup, nil
}
