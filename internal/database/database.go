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

// PrepareDB создает схему и таблицы для базы данных.
// Если произошла ошибка при создании схемы или таблиц, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании схемы и таблиц.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании схемы или таблиц.
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
	return nil
}

// CreateSchema создает схему для базы данных.
// Если произошла ошибка при создании схемы, программа завершается с кодом ошибки.
// В случае успеха, выводится сообщение об успешном создании схемы.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании схемы.
func CreateSchema() error {
	_, err := DB.Exec("CREATE SCHEMA IF NOT EXISTS loyalty")
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
    			role VARCHAR(255) DEFAULT 'user')
    			`
	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	config.Logger.Info("User table is ready")
	return nil
}
