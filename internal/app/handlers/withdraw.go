// Package handlers предоставляет функции для обработки HTTP-запросов в системе лояльности.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
	"github.com/FollowLille/loyalty/internal/utils"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

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

	var request WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		config.Logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if !utils.CheckLunar(request.Order) {
		config.Logger.Error("Failed to check lunar")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order"})
		return
	}

	currentBalance, _, err := database.FetchUserBalance(userIDInt)
	if err != nil {
		config.Logger.Error("Failed to fetch user balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user balance"})
		return
	}

	if currentBalance < request.Sum {
		config.Logger.Error("Failed to withdraw due to insufficient balance")
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	if err := database.RegisterWithdraw(request.Order, request.Sum); err != nil {
		config.Logger.Error("Failed to withdraw", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to withdraw"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})

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

	withdrawals, err := database.FetchUserWithdrawals(userIDInt)
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

	response := make([]WithdrawResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		response[i] = WithdrawResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}
