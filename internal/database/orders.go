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

	query := `INSERT INTO loyalty.orders (id, status) VALUES ($1, 1) RETURNING id`
	var orderID int
	row, err := QueryRowWithRetry(ctx, tx, query, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to get order ID: %w", err)
	}
	if err = row.Scan(&orderID); err != nil {
		return fmt.Errorf("failed to scan order ID: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
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
		SELECT o.id, sd.status_name, b.accrual, o.created_at
		FROM loyalty.orders o
		JOIN loyalty.user_orders uo ON o.id = uo.order_id
		JOIN loyalty.status_dictionary sd ON o.status = sd.id
		LEFT JOIN loyalty.bonuses b ON b.order_id = o.id
		WHERE uo.user_id = $1
		ORDER BY o.created_at DESC;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := QueryRowsWithRetry(ctx, DB, query, userID)
	if err != nil {
		config.Logger.Error("Failed to fetch user orders", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch user orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			config.Logger.Error("Failed to scan order", zap.Error(err))
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		config.Logger.Error("Failed to fetch user orders", zap.Error(rows.Err()))
		return nil, fmt.Errorf("failed to fetch user orders: %w", rows.Err())
	}

	return orders, nil
}

func GetOrdersByStatus() ([]Order, error) {
	var orders []Order
	query := `
			SELECT o.id, sd.status_name
			FROM loyalty.orders o
			JOIN loyalty.status_dictionary sd ON o.status = sd.id
			WHERE status_name IN ('NEW', 'PROCESSING');`
	rows, err := QueryRowsWithRetry(context.Background(), DB, query)
	if err != nil {
		config.Logger.Error("Failed to get orders by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.Number, &order.Status); err != nil {
			config.Logger.Error("Failed to scan order", zap.Error(err))
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		config.Logger.Error("Failed to get orders by status", zap.Error(rows.Err()))
		return nil, fmt.Errorf("failed to get orders by status: %w", rows.Err())
	}

	return orders, nil
}

func UpdateOrder(orderNumber, status string, accrual int64) error {
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
		UPDATE loyalty.orders 
		SET status = (
			SELECT id
			FROM loyalty.status_dictionary
			WHERE status_name = $1
			LIMIT 1
			) 
	  	WHERE id = $2;`
	err = ExecQueryWithRetry(ctx, tx, query, status, orderNumber)

	if err != nil {
		config.Logger.Error("Failed to update order", zap.Error(err), zap.String("query", query))
		return fmt.Errorf("failed to update order: %w", err)
	}

	query = `
		INSERT INTO loyalty.bonuses (order_id, accrual) 
		VALUES ($1, $2) 
		ON CONFLICT (order_id) 
		DO UPDATE SET accrual = EXCLUDED.accrual`
	err = ExecQueryWithRetry(ctx, tx, query, orderNumber, accrual)
	if err != nil {
		config.Logger.Error("Failed to update order", zap.Error(err), zap.String("query", query))
		return fmt.Errorf("failed to update order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		config.Logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
