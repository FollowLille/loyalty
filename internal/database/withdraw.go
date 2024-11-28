// Package database предоставляет функции для работы с базой данных в системе лояльности.
// Включает функции для работы с выводами баланса пользователя
package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
)

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

// RegisterWithdraw регистрирует вывод баланса пользователя
// Если произошла ошибка при выполнении запроса, программа завершается с кодом ошибки.
// В случае успеха, возвращает nil.
//
// Параметры:
//   - userID: идентификатор пользователя.
//   - orderNumber: идентификатор заказа.
//   - sum: сумма вывода.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func RegisterWithdraw(userID int64, orderNumber string, sum float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := DB.BeginTx(ctx, nil)

	if err != nil {
		config.Logger.Error("Failed to start transaction", zap.Error(err))
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				config.Logger.Error("Failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	query := `
		INSERT INTO loyalty.bonuses (order_id, withdrawn)
		VALUES ($1, $2)
		ON CONFLICT (order_id) 
		DO UPDATE SET withdrawn = EXCLUDED.withdrawn
		returning order_id
		`

	var orderID int
	row, err := QueryRowWithRetry(ctx, tx, query, orderNumber, sum)
	if err != nil {
		return fmt.Errorf("failed to get order ID: %w", err)
	}
	if err = row.Scan(&orderID); err != nil {
		return fmt.Errorf("failed to scan order ID: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	err = ExecQueryWithRetry(ctx, tx, `INSERT INTO loyalty.orders (id, status) VALUES ($1, 4) on conflict (id) do update set status = EXCLUDED.status`, orderID)
	if err != nil {
		config.Logger.Error("Failed to update order status", zap.Error(err))
		return fmt.Errorf("failed to update order status: %w", err)
	}

	err = ExecQueryWithRetry(ctx, tx, `INSERT INTO loyalty.user_orders (user_id, order_id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		return fmt.Errorf("failed to link user and order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FetchUserWithdrawals возвращает список выводов баланса пользователя
// Если произошла ошибка при выполнении запроса, программа завершается с кодом ошибки.
// В случае успеха, возвращает список выводов баланса пользователя.
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []Withdrawal: список выводов баланса пользователя.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func FetchUserWithdrawals(userID int64) ([]Withdrawal, error) {
	query := `
		SELECT 
		    o.id as order_number,
		    b.withdrawn as sum,
		    b.created_at as processed_at
		FROM loyalty.bonuses b 
		JOIN loyalty.orders o ON o.id = b.order_id
		JOIN loyalty.user_orders uo on o.id = uo.order_id
		WHERE uo.user_id = $1 and b.withdrawn > 0
		ORDER BY processed_at DESC;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var withdrawals []Withdrawal
	rows, err := QueryRowsWithRetry(ctx, DB, query, userID)
	if err != nil {
		config.Logger.Error("Failed to fetch user withdrawals", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal Withdrawal
		if err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			config.Logger.Error("Failed to scan row", zap.Error(err))
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if rows.Err() != nil {
		config.Logger.Error("Failed to fetch user withdrawals", zap.Error(rows.Err()))
		return nil, rows.Err()
	}

	return withdrawals, nil

}
