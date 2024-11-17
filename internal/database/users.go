// Package database предоставляет функции для работы в системе лояльности.
// Включает функции для проверки существования пользователя, создания нового пользователя и получения информации о пользователе.
// Все функции используют механизм повторных попыток при выполнении SQL-запросов.

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/FollowLille/loyalty/internal/config"
	cstmerr "github.com/FollowLille/loyalty/internal/errors"
)

// IsUserExists проверяет, существует ли пользователь с указанным именем в базе данных.
// Если пользователь существует, возвращает true. В противном случае возвращает false.
//
// Параметры:
//   - name: имя пользователя для проверки.
//
// Возвращает:
//   - bool: true, если пользователь существует; false в противном случае.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func IsUserExists(name string) (bool, error) {
	query := "SELECT EXISTS (SELECT 1 FROM loyalty.users WHERE name = $1)"
	var exists bool
	row, err := QueryRowWithRetry(context.Background(), DB, query, name)
	if err != nil {
		config.Logger.Error("Failed to check if user exists", zap.Error(err))
		return false, err
	}
	if err = row.Scan(&exists); err != nil {
		config.Logger.Error("Failed to scan result", zap.Error(err))
		return false, err
	}
	return exists, nil
}

// CreateUser создает нового пользователя в базе данных с указанными именем и хэшем пароля.
// Если пользователь с таким именем уже существует, возвращает ошибку cstmerr.ErrorUserAlreadyExists.
//
// Параметры:
//   - name: имя пользователя для создания.
//   - passwordHash: хэш пароля пользователя.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при создании пользователя или пользователь уже существует.
func CreateUser(name, passwordHash string) error {
	exists, err := IsUserExists(name)
	if err != nil {
		return err
	}

	if exists {
		config.Logger.Warn("User already exists", zap.String("user", name))
		return cstmerr.ErrorUserAlreadyExists
	}

	query := "INSERT INTO loyalty.users (name, password_hash) VALUES ($1, $2)"
	err = ExecQueryWithRetry(context.Background(), DB, query, name, passwordHash)
	if err != nil {
		config.Logger.Error("Failed to create user", zap.Error(err))
		return err
	}
	config.Logger.Info("User created", zap.String("user", name))
	return nil
}

// GetUserPasswordHash получает хэш пароля пользователя по его имени.
// Если пользователь не найден, возвращает пустую строку и nil.
//
// Параметры:
//   - name: имя пользователя для поиска.
//
// Возвращает:
//   - string: хэш пароля пользователя, если он найден.
//   - error: ошибка, если произошла ошибка при выполнении запроса.
func GetUserPasswordHash(name string) (string, error) {
	query := "SELECT password_hash FROM loyalty.users WHERE name = $1"
	var passwordHash string
	row, err := QueryRowWithRetry(context.Background(), DB, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			config.Logger.Warn("User does not exist", zap.String("user", name))
			return "", nil
		}
		config.Logger.Error("Failed to get user password hash", zap.Error(err))
		return "", err
	}
	err = row.Scan(&passwordHash)
	if err != nil {
		config.Logger.Error("Failed to scan result", zap.Error(err))
		return "", err
	}
	config.Logger.Info("User password hash retrieved", zap.String("user", name))
	return passwordHash, nil
}

func ValidateUser(username, password string) (int64, error) {
	var userID int64
	var hashedPassword string

	query := `SELECT id, password_hash FROM loyalty.users WHERE name = $1`
	err := DB.QueryRow(query, username).Scan(&userID, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		config.Logger.Error("Failed to validate user", zap.Error(err))
		return 0, fmt.Errorf("failed to validate user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return 0, err
	}

	return userID, nil
}

func GetUserIdByName(name string) (int64, error) {
	var userID int64
	query := `SELECT id FROM loyalty.users WHERE name = $1`
	row, err := QueryRowWithRetry(context.Background(), DB, query, name)
	if err != nil {
		return 0, err
	}
	err = row.Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		config.Logger.Error("User not found", zap.String("user", name))
	} else if err != nil {
		config.Logger.Error("Failed to scan result", zap.Error(err))
		return 0, err
	}
	return userID, nil
}
