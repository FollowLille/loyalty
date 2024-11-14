// Package database предоставляет функции для работы с базой данных в системе лояльности.
// Включает функции для работы с заказами пользователя
package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
)

type Order struct {
	Number     string
	Status     string
	Accrual    *int
	UploadedAt time.Time
}

// CreateOrder создает новый заказ для пользователя.
// Если произошла ошибка при создании заказа, программа завершается с кодом ошибки.
// В случае успеха, возвращает идентификатор созданного заказа.
//
// Параметры:
//   - userID: идентификатор пользователя.
//   - orderNumber: номер заказа.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании заказа.
func CreateOrder(userID int64, orderNumber string) error {
	ctx := context.Background()
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

	query := `INSERT INTO loyalty.orders (id, status) VALUES ($1, 1) RETURNING id`
	var orderID int
	err = QueryRowWithRetry(ctx, tx, query, orderNumber, orderID)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Создаем связь между пользователем и заказом
	err = ExecQueryWithRetry(ctx, tx, `INSERT INTO loyalty.user_orders (user_id, order_id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		return fmt.Errorf("failed to link user and order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserOrders возвращает информацию о заказах пользователя
//
// Параметры:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []Order: информация о заказах пользователя.
//   - error: ошибка, если произошла ошибка при получении информации о заказах пользователя.
func GetUserOrders(userID int64) ([]Order, error) {
	query := `
		SELECT o.id, sd.status, b.accrual, o.created_at
		FROM loyalty.orders o
		JOIN loyalty.user_orders uo ON o.id = uo.order_id
		JOIN loyalty.status_dictionary sd ON o.status = sd.id
		LEFT JOIN loyalty.bonuses b ON b.order_id = o.id
		WHERE uo.user_id = $1
		ORDER BY o.created_at DESC;
	`

	ctx := context.Background()
	rows, err := QueryRowsWithRetry(ctx, DB, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
