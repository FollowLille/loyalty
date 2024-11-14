// Package database предоставляет функции для работы с базой данных в системе лояльности.
// Включает функции для работы с балансом пользователя
package database

import (
	"context"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
)

// FetchUserBalance возвращает текущий баланс пользователя и общую сумму его выводов
// Если произошла ошибка при выполнении запроса, программа завершается с кодом ошибки.
// В случае успеха, возвращает текущий баланс пользователя и общую сумму его выводов.
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - float64: текущий баланс пользователя.
//   - float64: общую сумму его выводов.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func FetchUserBalance(userID int64) (float64, float64, error) {
	query := `
		SELECT
			COALESCE(ub.total_accruals - ub.total_withdrawn, 0) as current_balance,
			COALESCE(ub.total_withdrawn, 0) as total_withdrawn
		FROM loyalty.user_bonuses ub
		WHERE ub.user_id = $1;
	`

	var currentBalance float64
	var totalWithdrawn float64

	err := QueryRowWithRetry(context.Background(), DB, query, userID, &currentBalance, &totalWithdrawn)
	if err != nil {
		config.Logger.Error("Failed to fetch user balance", zap.Error(err))
		return 0, 0, err
	}
	return currentBalance, totalWithdrawn, nil
}
