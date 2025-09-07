package mongo

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fakeInserter struct {
	calls int
	err   error
}

func (f *fakeInserter) InsertOne(ctx context.Context, doc interface{}, _ ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	f.calls++
	return nil, f.err
}

func TestInsertWithRetry_NetworkErrorRetries(t *testing.T) {
	t.Cleanup(func() { isNetworkErr = mongo.IsNetworkError })
	isNetworkErr = func(error) bool { return true }
	fi := &fakeInserter{err: errors.New("net err")}
	l := &Logger{collection: fi}
	ev := AuthLog{EventType: "login", UserID: "u"}
	_ = l.insertWithRetry(context.Background(), ev)
	if fi.calls < 2 {
		t.Fatalf("expected retry on network error, got %d calls", fi.calls)
	}
}

func TestInsertWithRetry_NonNetworkNoRetry(t *testing.T) {
	t.Cleanup(func() { isNetworkErr = mongo.IsNetworkError })
	isNetworkErr = func(error) bool { return false }
	fi := &fakeInserter{err: errors.New("server error")}
	l := &Logger{collection: fi}
	ev := AuthLog{EventType: "login", UserID: "u"}
	_ = l.insertWithRetry(context.Background(), ev)
	if fi.calls != 1 {
		t.Fatalf("expected no retry on non-network error, got %d calls", fi.calls)
	}
}
