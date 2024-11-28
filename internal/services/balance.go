package services

import (
	"errors"

	"github.com/FollowLille/loyalty/internal/database"
)

type UserBalance struct {
	Current   float64
	Withdrawn float64
}

// FetchUserBalance выполняет бизнес-логику для получения баланса пользователя.
// Если произошла ошибка при выполнении запроса, возвращает ошибку.
// В случае успеха, возвращает баланс пользователя.
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращаемое значение:
//   - balance: баланс пользователя.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func FetchUserBalance(userID int64) (UserBalance, error) {
	balance, withdrawn, err := database.FetchUserBalance(userID)
	if err != nil {
		return UserBalance{}, errors.New("failed to fetch user balance")
	}

	return UserBalance{
		Current:   balance,
		Withdrawn: withdrawn,
	}, nil
}
