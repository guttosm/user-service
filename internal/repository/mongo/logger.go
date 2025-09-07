package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthLog struct {
	EventType string    `bson:"event_type"`
	UserID    string    `bson:"user_id"`
	Timestamp time.Time `bson:"timestamp"`
	IP        string    `bson:"ip,omitempty"`
	Metadata  any       `bson:"metadata,omitempty"`
}

type mongoInserter interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

type Logger struct{ collection mongoInserter }

// indirection for testability
var isNetworkErr = mongo.IsNetworkError

func NewLogger(db *mongo.Database, collectionName string) *Logger {
	return &Logger{
		collection: db.Collection(collectionName),
	}
}

// NewTestLogger returns a Logger with a custom inserter. For tests only.
func NewTestLogger(inserter mongoInserter) *Logger { return &Logger{collection: inserter} }

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
		if err == nil || !isNetworkErr(err) {
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
