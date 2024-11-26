// Package services предоставляет бизнес часть для работы с пользователями в системе лояльности.
package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/FollowLille/loyalty/internal/auth"
	"github.com/FollowLille/loyalty/internal/database"
)

// RegisterUser регистрирует нового пользователя в системе лояльности.
// Если пользователь существует, возвращает ошибку cstmerr.ErrorUserAlreadyExists.
// Если пользователь не существует, создает нового пользователя и возвращает токен.
//
// Параметры:
//   - username: имя пользователя.
//   - password: пароль пользователя.
//
// Возвращаемое значение:
//   - token: токен для доступа к системе лояльности.
//   - error: ошибка, если произошла ошибка при регистрации пользователя.
func RegisterUser(username, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	err = database.CreateUser(username, string(hashedPassword))
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(username)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

// LoginUser выполняет вход пользователя в систему лояльности.
// Если пользователь существует, создает новый токен и возвращает его в качестве ответа.
//
// Параметры:
//   - username: имя пользователя.
//   - password: пароль пользователя.
//
// Возвращаемое значение:
//   - token: токен для доступа к системе лояльности.
//   - error: ошибка, если произошла ошибка при входе пользователя.
func LoginUser(username, password string) (string, error) {
	storedHash, err := database.GetUserPasswordHash(username)
	if err != nil {
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token, err := auth.GenerateToken(username)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
