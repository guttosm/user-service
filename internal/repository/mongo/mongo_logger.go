package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuthLog struct {
	EventType string    `bson:"event_type"`
	UserID    string    `bson:"user_id"`
	Timestamp time.Time `bson:"timestamp"`
	IP        string    `bson:"ip,omitempty"`
	Metadata  any       `bson:"metadata,omitempty"`
}

type Logger struct {
	collection *mongo.Collection
}

func NewLogger(db *mongo.Database, collectionName string) *Logger {
	return &Logger{
		collection: db.Collection(collectionName),
	}
}

// validateAuthLog checks required fields.
func validateAuthLog(event AuthLog) error {
	if event.EventType == "" {
		return errors.New("event_type is required")
	}
	if event.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// insertWithRetry tries to insert the log up to 3 times for transient errors.
func (l *Logger) insertWithRetry(ctx context.Context, event AuthLog) error {
	var err error
	for i := 0; i < 3; i++ {
		_, err = l.collection.InsertOne(ctx, event)
		if err == nil || !mongo.IsNetworkError(err) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

// Log inserts a new authentication event into MongoDB with validation and context.
func (l *Logger) Log(ctx context.Context, event AuthLog) error {
	if err := validateAuthLog(event); err != nil {
		return err
	}
	event.Timestamp = time.Now()
	return l.insertWithRetry(ctx, event)
}
