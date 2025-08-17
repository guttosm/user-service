//go:build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startMongo(t *testing.T) (*mongodrv.Client, func()) {
	t.Helper()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},

		Cmd: []string{"mongod", "--bind_ip_all", "--setParameter", "enableTestCommands=1"},
		WaitingFor: wait.ForListeningPort("27017/tcp").
			WithStartupTimeout(90 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "27017/tcp")
	require.NoError(t, err)

	uri := "mongodb://" + host + ":" + port.Port()
	client, err := mongodrv.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.Ping(ctx, nil) == nil
	}, 20*time.Second, 200*time.Millisecond)

	cleanup := func() {
		_ = client.Disconnect(context.Background())
		_ = c.Terminate(context.Background())
	}
	return client, cleanup
}

func TestLogger_Integration(t *testing.T) {
	t.Parallel()

	client, cleanup := startMongo(t)
	defer cleanup()

	db := client.Database("user_service_test")
	col := "auth_logs"
	logger := NewLogger(db, col)

	type in struct {
		event AuthLog
		setup func(t *testing.T)
	}
	type out struct {
		wantErr        bool
		expectInserted bool
	}

	tests := []struct {
		name string
		in   in
		out  out
	}{
		{
			name: "happy path: insert ok and timestamp set",
			in: in{
				event: AuthLog{
					EventType: "login_success",
					UserID:    "u-123",
					IP:        "127.0.0.1",
					Metadata:  map[string]any{"agent": "test"},
				},
				setup: func(t *testing.T) {
					_ = db.Collection(col).Drop(context.Background())
				},
			},
			out: out{wantErr: false, expectInserted: true},
		},
		{
			name: "validation error (missing fields): should not insert",
			in: in{
				event: AuthLog{
					EventType: "",
					UserID:    "",
				},
				setup: func(t *testing.T) {
					_ = db.Collection(col).Drop(context.Background())
				},
			},
			out: out{wantErr: true, expectInserted: false},
		},
		{
			name: "retry once on transient network error via failpoint",
			in: in{
				event: AuthLog{
					EventType: "login_failure",
					UserID:    "u-999",
				},
				setup: func(t *testing.T) {
					_ = db.Collection(col).Drop(context.Background())

					admin := client.Database("admin")
					cmd := bson.D{
						{"configureFailPoint", "failCommand"},
						{"mode", bson.D{{"times", 1}}},
						{"data", bson.D{
							{"failCommands", bson.A{"insert"}},
							{"closeConnection", true},
						}},
					}
					var res bson.M
					err := admin.RunCommand(context.Background(), cmd).Decode(&res)
					require.NoError(t, err, "enabling failpoint should succeed")
				},
			},
			out: out{wantErr: false, expectInserted: true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.in.setup != nil {
				tc.in.setup(t)
			}

			start := time.Now()
			err := logger.Log(context.Background(), tc.in.event)

			if tc.out.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.out.expectInserted {
				filter := bson.D{
					{"user_id", tc.in.event.UserID},
					{"event_type", tc.in.event.EventType},
				}
				var got AuthLog
				err := db.Collection(col).FindOne(context.Background(), filter).Decode(&got)
				require.NoError(t, err, "inserted document should be found")

				assert.Equal(t, tc.in.event.UserID, got.UserID)
				assert.Equal(t, tc.in.event.EventType, got.EventType)
				assert.WithinDuration(t, start, got.Timestamp, 5*time.Second)
			}
		})
	}
}
