// Package handlers предоставляет функции для обработки HTTP-запросов в системе лояльности.
package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/services"
)

// GetWithdrawRequest обрабатывает запрос на вывод баланса пользователя.
// Если произошла ошибка при выполнении запроса, программа завершается с кодом ошибки.
// В случае успеха, возвращает nil.
//
// Параметры:
//   - c: контекст HTTP-запроса.
func GetWithdrawRequest(c *gin.Context) {
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

	var request services.WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		config.Logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := services.ProcessWithdrawRequest(userIDInt, request); err != nil {
		if err.Error() == "invalid order number" {
			config.Logger.Error("Failed to process withdraw due to invalid order number")
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order"})
			return
		} else if err.Error() == "insufficient balance" {
			config.Logger.Error("Failed to process withdraw due to insufficient balance")
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
			return
		}

		config.Logger.Error("Failed to process withdraw", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdraw"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdraw request processed"})
}

// GetWithdrawals возвращает список выводов баланса пользователя.
//
// Параметры:
//   - c: контекст HTTP-запроса.
func GetWithdrawals(c *gin.Context) {
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

	withdrawals, err := services.FetchWithdrawals(userIDInt)
	if err != nil {
		config.Logger.Error("Failed to fetch user withdrawals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user withdrawals"})
		return
	}

	if len(withdrawals) == 0 {
		config.Logger.Info("No withdrawals found")
		c.JSON(http.StatusNoContent, gin.H{"message": "No withdrawals found"})
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}
