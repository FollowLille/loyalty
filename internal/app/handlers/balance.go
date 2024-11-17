// Package handlers предоставляет функции для работы с накопительным счётом пользователя.
// Включает в себя функции для обработки запросов связанных с балансом пользователя
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
)

// BalanceResponse представляет информацию о балансе пользователя
type BalanceResponse struct {
	CurrentBalance float64 `json:"current_balance"`
	TotalWithdrawn float64 `json:"total_withdrawn"`
}

// GetUserBalance возвращает информацию о балансе пользователя
//
// Параметры:
//   - c: контекст HTTP-запроса.
func GetBalance(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		config.Logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id is not a string"})
		return
	}

	userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id format"})
		return
	}

	balance, withdrawn, err := database.FetchUserBalance(userIDInt)
	if err != nil {
		config.Logger.Error("Failed to fetch user balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user balance"})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		CurrentBalance: balance,
		TotalWithdrawn: withdrawn,
	})
}
