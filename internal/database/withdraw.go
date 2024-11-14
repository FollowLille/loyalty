// Package database предоставляет функции для работы с базой данных в системе лояльности.
// Включает функции для работы с выводами баланса пользователя
package database

import (
	"context"
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
//   - orderNumber: идентификатор заказа.
//   - sum: сумма вывода.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func RegisterWithdraw(orderNumber string, sum float64) error {
	query := `
		INSERT INTO loyalty.bonuses (order_id, withdrawn)
		VALUES ($1, $2)
		`

	err := ExecQueryWithRetry(context.Background(), DB, query, orderNumber, sum)
	if err != nil {
		config.Logger.Error("Failed to register withdraw", zap.Error(err))
		return err
	}

	return nil
}

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

	ctx := context.Background()
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

	return withdrawals, nil

}
