package retry

import (
	"github.com/FollowLille/loyalty/internal/config"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func Retry(f func() error) error {

	var err error
	for _, delay := range config.DatabaseRetryDelays {
		err = f()
		if err == nil {
			return nil
		}
		if err == ErrorNonRetriable || err == ErrorNonRetriablePostgres {
			return err
		}
		config.Logger.Info("Retrying after delay", zap.Duration("delay", delay))
		time.Sleep(delay)
	}

	return err
}

func IsRetriablePostgresError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionFailure,
			pgerrcode.AdminShutdown,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected:
			return true
		}
	}
	return false
}
