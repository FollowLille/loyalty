package services

import (
	"errors"
	"time"

	"github.com/FollowLille/loyalty/internal/database"
	"github.com/FollowLille/loyalty/internal/utils"
)

type Order struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

// GetOrders выполняет бизнес-логику для получения списка заказов пользователя.
// Если произошла ошибка при выполнении запроса, возвращает ошибку.
// В случае успеха, возвращает список заказов пользователя.
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращаемое значение:
//   - orders: список заказов пользователя.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func GetOrders(userID int64) ([]map[string]interface{}, error) {
	orders, err := database.GetUserOrders(userID)
	if err != nil {
		return nil, errors.New("failed to fetch orders")
	}

	if len(orders) == 0 {
		return nil, nil
	}

	response := make([]map[string]interface{}, len(orders))
	for i, order := range orders {
		response[i] = map[string]interface{}{
			"number":      order.Number,
			"status":      order.Status,
			"accrual":     order.Accrual,
			"uploaded_at": order.UploadedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}

// UploadOrder выполняет бизнес-логику для загрузки заказа пользователя.
// Если произошла ошибка при выполнении запроса, возвращает ошибку.
//
// Параметры:
//   - userID: идентификатор пользователя.
//   - orderNumber: номер заказа.
//
// Возвращаемое значение:
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func UploadOrder(userID int64, orderNumber string) error {
	if !utils.CheckLunar(orderNumber) {
		return errors.New("invalid order number")
	}

	ownerID, err := database.GetOrderOwner(orderNumber)
	if err != nil {
		return errors.New("failed to get order owner")
	}

	if ownerID != nil {
		if *ownerID == userID {
			return errors.New("order already uploaded by you")
		}
		return errors.New("order already uploaded by another user")
	}

	if err := database.CreateOrder(userID, orderNumber); err != nil {
		return errors.New("failed to create order")
	}

	return nil
}
