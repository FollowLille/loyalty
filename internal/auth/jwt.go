// Package auth предоставляет функции для работы с JWT-токенами
package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/FollowLille/loyalty/internal/config"
)

// GenerateToken создает JWT-токен для указанного пользователя.
// JWT-токен содержит имя пользователя и время его действия.
// Время действия токена устанавливается на 24 часа.
// Токен сгенерирован и возвращается в виде строки.
// Параметры:
//   - username: имя пользователя.
//
// Возвращает:
//   - string: JWT-токен для указанного пользователя.
//   - error: ошибка, если произошла ошибка при генерации токена.
func GenerateToken(username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("empty username")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	secretKey := []byte(config.SuperSecretKey)
	return token.SignedString(secretKey)
}

// ValidateToken проверяет JWT-токен на валидность.
// JWT-токен должен содержать имя пользователя и время его действия.
// Если токен валиден, функция возвращает имя пользователя.
// Если токен невалиден, функция возвращает ошибку.
// Параметры:
//   - tokenStr: JWT-токен для проверки.
//
// Возвращает:
//   - string: имя пользователя, если токен валиден.
//   - error: ошибка, если произошла ошибка при проверке токена.
func ValidateToken(tokenStr string) (string, error) {
	// Разбор и валидация токена
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что используется HMAC с алгоритмом HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Возвращаем ключ в виде []byte
		return []byte(config.SuperSecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	// Проверка валидности токена и извлечение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Проверяем expiration (exp)
	expirationTime, ok := claims["exp"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid expiration time format")
	}
	if time.Now().Unix() > int64(expirationTime) {
		return "", fmt.Errorf("token expired")
	}

	// Извлекаем username
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return "", fmt.Errorf("invalid or missing username in token")
	}

	return username, nil
}
