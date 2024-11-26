package services

import (
	"errors"
	"time"

	"github.com/FollowLille/loyalty/internal/database"
	"github.com/FollowLille/loyalty/internal/utils"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// ProcessWithdrawRequest выполняет бизнес-логику для обработки запроса на вывод средств.
// Если произошла ошибка при выполнении запроса, возвращает ошибку.
//
// Параметры:
//   - userID: идентификатор пользователя.
//   - req: запрос на вывод средств.
//
// Возвращаемое значение:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func ProcessWithdrawRequest(userID int64, req WithdrawRequest) error {
	if !utils.CheckLunar(req.Order) {
		return errors.New("invalid order number")
	}

	currentBalance, _, err := database.FetchUserBalance(userID)
	if err != nil {
		return errors.New("failed to fetch user balance")
	}

	if currentBalance < req.Sum {
		return errors.New("insufficient balance")
	}

	if err := database.RegisterWithdraw(req.Order, req.Sum); err != nil {
		return errors.New("failed to register withdrawal")
	}

	return nil
}

// FetchWithdrawals выполняет бизнес-логику для получения списка выводов пользователя.
// Если произошла ошибка при выполнении запроса, возвращает ошибку.
// В случае успеха, возвращает список выводов пользователя.
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращаемое значение:
//   - withdrawals: список выводов пользователя.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func FetchWithdrawals(userID int64) ([]WithdrawResponse, error) {
	withdrawals, err := database.FetchUserWithdrawals(userID)
	if err != nil {
		return nil, errors.New("failed to fetch user withdrawals")
	}

	response := make([]WithdrawResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		response[i] = WithdrawResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		}
	}

	return response, nil
}
