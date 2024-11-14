// Package database предоставляет функции для работы с базой данных в системе лояльности.
// Включает функции для подключения к базе данных, создания схемы и таблиц, инициализации базы данных.
package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
)

// DB хранит глобальное соединение с базой данных.
var DB *sql.DB

type DBHandler interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// InitDB инициализирует соединение с базой данных.
// Если произошла ошибка при инициализации, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном подключении к базе данных.
//
// Параметры:
//   - connStr: строка подключения к базе данных.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при инициализации базы данных.
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		config.Logger.Fatal("Failed to connect to database", zap.Error(err))
		return err
	}

	err = DB.Ping()
	if err != nil {
		config.Logger.Fatal("Failed to ping database", zap.Error(err))
		return err
	}

	config.Logger.Info("Database connected")
	return nil
}

// PrepareDB создает схему, таблицы и VIEW для базы данных.
func PrepareDB() error {
	var err error
	if err = CreateSchema(); err != nil {
		config.Logger.Fatal("Failed to create schema", zap.Error(err))
		return err
	}

	if err = CreateUserTable(); err != nil {
		config.Logger.Fatal("Failed to create user table", zap.Error(err))
		return err
	}

	if err = CreateOrdersTable(); err != nil {
		config.Logger.Fatal("Failed to create orders table", zap.Error(err))
		return err
	}

	if err = CreateStatusDictionary(); err != nil {
		config.Logger.Fatal("Failed to create status dictionary", zap.Error(err))
		return err
	}

	if err = CreateBonusesTable(); err != nil {
		config.Logger.Fatal("Failed to create bonuses table", zap.Error(err))
		return err
	}

	if err = CreateUsersOrdersTable(); err != nil {
		config.Logger.Fatal("Failed to create user orders table", zap.Error(err))
		return err
	}

	if err = CreateOrdersBonusesTable(); err != nil {
		config.Logger.Fatal("Failed to create orders bonuses table", zap.Error(err))
		return err
	}

	// Создание VIEW для подсчета бонусов
	if err = CreateUserBonusesView(); err != nil {
		config.Logger.Fatal("Failed to create user bonuses view", zap.Error(err))
		return err
	}

	return nil
}

// CreateSchema создает схему для базы данных.
// Если произошла ошибка при создании схемы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании схемы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании схемы.
func CreateSchema() error {
	_, err := DB.Exec("CREATE SCHEMA IF NOT EXISTS loyalty;")
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	config.Logger.Info("Schema is ready")
	return nil
}

// CreateUserTable создает таблицу для хранения информации о пользователях.
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateUserTable() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.users (
    			id SERIAL PRIMARY KEY NOT NULL, 
    			name VARCHAR(255) NOT NULL, 
    			password_hash VARCHAR(255) NOT NULL, 
    			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
    			`
	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create user table", zap.Error(err))
		return fmt.Errorf("failed to create user table: %w", err)
	}
	config.Logger.Info("User table is ready")
	return nil
}

// CreateOrdersTable создает таблицу для хранения информации о заказах.
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateOrdersTable() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.orders (
				id SERIAL PRIMARY KEY NOT NULL,
				status INT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
			`

	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create orders table", zap.Error(err))
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	config.Logger.Info("Orders table is ready")
	return nil
}

// CreateStatusDictionary создает таблицу для хранения информации о статусах заказа.
// Так же мы сразу заполняем таблицу нужными статусами
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateStatusDictionary() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.status_dictionary (
				id INT PRIMARY KEY NOT NULL,
				status VARCHAR(255) NOT NULL,
				is_closed BOOLEAN NOT NULL,
				CONSTRAINT unique_status UNIQUE (status));
			`

	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create status dictionary table", zap.Error(err))
		return fmt.Errorf("failed to create status dictionary table: %w", err)
	}
	config.Logger.Info("Status dictionary table is ready")

	insert_query := `
		INSERT INTO loyalty.status_dictionary (id, status, is_closed) VALUES
		  (1, 'NEW', false), (2, 'PROCESSING', false), (3, 'INVALID', false), (4, 'PROCESSED', true)
		ON CONFLICT (status) DO NOTHING;`

	_, err = DB.Exec(insert_query)
	if err != nil {
		config.Logger.Error("Failed to create status dictionary table", zap.Error(err))
		return fmt.Errorf("failed to create status dictionary table: %w", err)
	}
	config.Logger.Info("Status dictionary table is ready")
	return nil
}

// CreateBonusesTable создает таблицу для хранения информации о бонусах.
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateBonusesTable() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.bonuses (
				id SERIAL PRIMARY KEY NOT NULL,
				order_id INT NOT NULL,
				accrual INT NOT NULL DEFAULT 0,
				withdrawn INT NOT NULL DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
			    `
	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create bonuses table", zap.Error(err))
		return fmt.Errorf("failed to create bonuses table: %w", err)
	}
	config.Logger.Info("Bonuses table is ready")
	return nil
}

// CreateUserOrdersTable создает таблицу для хранения связи между пользователями и заказами
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateUsersOrdersTable() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.user_orders (
				order_id INT NOT NULL,
				user_id INT NOT NULL,
				FOREIGN KEY (order_id) REFERENCES loyalty.orders(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES loyalty.users(id) ON DELETE CASCADE);
			`

	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create user orders table", zap.Error(err))
		return fmt.Errorf("failed to create user orders table: %w", err)
	}
	config.Logger.Info("User orders table is ready")
	return nil
}

// CreateOrdersBonusesTable создает таблицу для хранения связи между заказами и бонусами
// Если произошла ошибка при создании таблицы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании таблицы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании таблицы.
func CreateOrdersBonusesTable() error {
	query := `
			CREATE TABLE IF NOT EXISTS loyalty.orders_bonuses (
				order_id INT NOT NULL,
				bonus_id INT NOT NULL,
				FOREIGN KEY (order_id) REFERENCES loyalty.orders(id) ON DELETE CASCADE,
				FOREIGN KEY (bonus_id) REFERENCES loyalty.bonuses(id) ON DELETE CASCADE);
			`

	_, err := DB.Exec(query)
	if err != nil {
		config.Logger.Error("Failed to create orders bonuses table", zap.Error(err))
		return fmt.Errorf("failed to create orders bonuses table: %w", err)
	}
	config.Logger.Info("Orders bonuses table is ready")
	return nil
}

// CreateUserBonusesView создает представление для хранения информации о текущем состоянии бонусов пользователей
// Если произошла ошибка при создании представления, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании представления.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании представления.
func CreateUserBonusesView() error {
	query := `
		CREATE OR REPLACE VIEW loyalty.user_bonuses AS
		SELECT
			u.id AS user_id,
			u.name AS user_name,
			COALESCE(SUM(CASE WHEN sd.is_closed = TRUE THEN b.accrual ELSE 0 END), 0) AS total_accruals,  -- Сумма начислений только для закрытых заказов
			COALESCE(SUM(CASE WHEN sd.status != 'INVALID' THEN b.withdrawn ELSE 0 END), 0) AS total_withdrawn -- Сумма списаний для всех заказов
		FROM
			loyalty.users u
		LEFT JOIN
			loyalty.user_orders uo ON uo.user_id = u.id
		LEFT JOIN
			loyalty.orders o ON o.id = uo.order_id
		LEFT JOIN
			loyalty.bonuses b ON b.order_id = o.id
		LEFT JOIN
			loyalty.status_dictionary sd ON sd.id = o.status
		GROUP BY
			u.id;
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user bonuses view: %w", err)
	}

	config.Logger.Info("User bonuses view is ready")
	return nil
}

// GetOrderOwner возвращает идентификатор пользователя, создавшего заказ с указанным номером
// Если произошла ошибка при выполнении запроса, программа завершается с кодом ошибки.
// В случае успеха, возвращается идентификатор пользователя.
//
// Параметры:
//   - orderNumber: номер заказа.
//
// Возвращает:
//   - *int64: идентификатор пользователя, создавшего заказ.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func GetOrderOwner(orderNumber string) (*int64, error) {
	query := `
		SELECT uo.user_id
		FROM loyalty.orders o
		JOIN loyalty.user_orders uo ON uo.order_id = o.id
		WHERE o.id = $1;
		`

	var userID *int64
	err := QueryRowWithRetry(context.Background(), DB, query, orderNumber, &userID)
	if err != nil {
		config.Logger.Error("Failed to get order owner", zap.Error(err))
		return nil, err
	}
	return userID, nil
}
