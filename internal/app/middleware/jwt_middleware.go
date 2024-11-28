// Package middleware предоставляет функции для обработки запросов на взаимодействие с программой лояльности
// Включает в себя функции для проверки JWT-токена
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/FollowLille/loyalty/internal/auth"
	"github.com/FollowLille/loyalty/internal/database"
)

// AuthMiddleware проверяет JWT-токен перед обработкой запросами
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if len(tokenString) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		userName, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		userID, err := database.GetUserIDByName(userName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		userIDStr := strconv.FormatInt(userID, 10)
		c.Set("user_id", userIDStr)
		c.Next()
	}
}
