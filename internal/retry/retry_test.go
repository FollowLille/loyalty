package retry

import (
	"errors"
	"github.com/FollowLille/loyalty/internal/config"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	cstmerr "github.com/FollowLille/loyalty/internal/errors"
)

var mockLogger, _ = zap.NewDevelopment()

func init() {
	config.Logger = mockLogger
	config.DatabaseRetryDelays = []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
}

func TestRetry(t *testing.T) {
	type args struct {
		f func() error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success_on_first_try",
			args: args{
				f: func() error { return nil },
			},
			wantErr: nil,
		},
		{
			name: "success_after_retry",
			args: args{
				f: func() func() error {
					attempts := 0
					return func() error {
						attempts++
						if attempts == 2 {
							return nil
						}
						return errors.New("error")
					}
				}(),
			},
			wantErr: nil,
		},
		{
			name: "non_retriable_error",
			args: args{
				f: func() error { return cstmerr.ErrorNonRetriable },
			},
			wantErr: cstmerr.ErrorNonRetriable,
		},
		{
			name: "non_retriable_postgres_error",
			args: args{
				f: func() error { return cstmerr.ErrorNonRetriablePostgres },
			},
			wantErr: cstmerr.ErrorNonRetriablePostgres,
		},
		{
			name: "all_retries_fail",
			args: args{
				f: func() error { return errors.New("error") },
			},
			wantErr: errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := Retry(tt.args.f)
			if tt.wantErr != nil {
				assert.EqualError(t, gotErr, tt.wantErr.Error(), "Retry() failed for test case: %v", tt.name)
			} else {
				assert.NoError(t, gotErr, "Retry() failed for test case: %v", tt.name)
			}
		})
	}
}

func TestIsRetriablePostgresError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "retriable_connection_failure",
			args: args{
				err: &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			},
			want: true,
		},
		{
			name: "retriable_deadlock_detected",
			args: args{
				err: &pgconn.PgError{Code: pgerrcode.DeadlockDetected},
			},
			want: true,
		},
		{
			name: "non_retriable_error",
			args: args{
				err: &pgconn.PgError{Code: pgerrcode.SyntaxError},
			},
			want: false,
		},
		{
			name: "non_pg_error",
			args: args{
				err: errors.New("some random error"),
			},
			want: false,
		},
		{
			name: "nil_error",
			args: args{
				err: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetriablePostgresError(tt.args.err)
			assert.Equal(t, tt.want, got, "IsRetriablePostgresError() failed for test case: %v", tt.name)
		})
	}
}
