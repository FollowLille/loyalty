// Package database предоставляет функции для повторного выполнения SQL-запросов.
// Все функции используют механизм повторных попыток при выполнении SQL-запросов.
package database

import (
	"context"
	"database/sql"
	"github.com/FollowLille/loyalty/internal/config"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	cstmerr "github.com/FollowLille/loyalty/internal/errors"
	"github.com/FollowLille/loyalty/internal/retry"
)

// ExecContexter представляет интерфейс для выполнения SQL-запросов.
// Используется как для sql.DB, так и для sql.Tx.
type ExecContexter interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// ExecQueryWithRetry выполняет SQL-запрос, не возвращающий результат, с повторными попытками при возникновении ошибок.
//
// Параметры:
//   - ctx: контекст для отмены операции.
//   - db: соединение с базой данных.
//   - query: SQL-запрос для выполнения.
//   - args: аргументы для SQL-запроса.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func ExecQueryWithRetry(ctx context.Context, db ExecContexter, query string, args ...interface{}) error {
	err := retry.Retry(func() error {
		_, execErr := db.ExecContext(ctx, query, args...)
		if execErr != nil {
			if retry.IsRetriablePostgresError(execErr) {
				config.Logger.Error("Retrying query")
				return cstmerr.ErrorRetriablePostgres
			}
			config.Logger.Error("Non retriable error during query execution", zap.Error(execErr))
			return cstmerr.ErrorNonRetriablePostgres
		}
		return nil
	})

	if err != nil {
		config.Logger.Error("Failed to execute query", zap.Error(err))
		return err
	}
	config.Logger.Info("Query executed successfully")
	return nil
}

// QueryRowWithRetry выполняет SQL-запрос, возвращающий одну строку, с повторными попытками при возникновении ошибок.
//
// Параметры:
//   - ctx: контекст для отмены операции.
//   - db: соединение с базой данных.
//   - query: SQL-запрос для выполнения.
//   - dest: указатель на переменную, в которую будет записан результат выполнения запроса.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func QueryRowWithRetry(ctx context.Context, db *sql.DB, query string, dest ...interface{}) error {
	err := retry.Retry(func() error {
		err := db.QueryRowContext(ctx, query).Scan(dest...)
		if err != nil {
			if retry.IsRetriablePostgresError(err) {
				config.Logger.Error("Retrying query")
				return cstmerr.ErrorRetriablePostgres
			}
			config.Logger.Error("Non retriable error during query execution", zap.Error(err))
			return cstmerr.ErrorNonRetriablePostgres
		}
		return nil
	})

	if err != nil {
		config.Logger.Error("Failed to execute query", zap.Error(err))
		return err
	}
	config.Logger.Info("Query executed successfully")
	return nil
}

// QueryRowsWithRetry выполняет SQL-запрос, возвращающий список строк, с повторными попытками при возникновении ошибок.
//
// Параметры:
//   - ctx: контекст для отмены операции.
//   - db: соединение с базой данных.
//   - query: SQL-запрос для выполнения.
//   - args: аргументы для SQL-запроса.
//
// Возвращает:
//   - *sql.Rows: указатель на результат выполнения запроса.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func QueryRowsWithRetry(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	err = retry.Retry(func() error {
		rows, err = db.QueryContext(ctx, query, args...)
		if err != nil {
			if retry.IsRetriablePostgresError(err) {
				config.Logger.Error("Retrying query")
				return cstmerr.ErrorRetriablePostgres
			}
			config.Logger.Error("Non retriable error during query execution", zap.Error(err))
			return cstmerr.ErrorNonRetriablePostgres
		}
		if rows.Err() != nil {
			config.Logger.Error("Failed to execute query", zap.Error(err))
			return cstmerr.ErrorNonRetriablePostgres
		}
		return nil
	})

	if err != nil {
		config.Logger.Error("Failed to execute query", zap.Error(err))
		return nil, err
	}
	config.Logger.Info("Query executed successfully")
	return rows, nil
}
