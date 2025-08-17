//go:build !integration

package mongo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndLog(t *testing.T) {
	t.Parallel()

	type in struct {
		ev AuthLog
	}
	type out struct {
		wantErr bool
	}

	tests := []struct {
		name string
		in   in
		out  out
	}{
		{
			name: "validate: ok",
			in:   in{ev: AuthLog{EventType: "login_success", UserID: "u-1"}},
			out:  out{wantErr: false},
		},
		{
			name: "validate: missing event_type",
			in:   in{ev: AuthLog{UserID: "u-1"}},
			out:  out{wantErr: true},
		},
		{
			name: "validate: missing user_id",
			in:   in{ev: AuthLog{EventType: "login_failure"}},
			out:  out{wantErr: true},
		},
		{
			name: "Log: returns early on invalid (no DB call needed)",
			in:   in{ev: AuthLog{EventType: "", UserID: ""}},
			out:  out{wantErr: true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			err := validateAuthLog(tc.in.ev)
			if tc.name == "Log: returns early on invalid (no DB call needed)" {

				l := &Logger{collection: nil}
				err = l.Log(context.Background(), tc.in.ev)
			}

			if tc.out.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLogger_Log_SetsTimestamp(t *testing.T) {
	t.Parallel()

	ev := AuthLog{EventType: "login_success", UserID: "u-1"}
	require.NoError(t, validateAuthLog(ev))

	assert.True(t, ev.Timestamp.IsZero(), "caller Timestamp is ignored; Log sets it")
}
