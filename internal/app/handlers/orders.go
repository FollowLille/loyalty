// Package handlers предоставляет функции для обработки HTTP-запросов в системе лояльности.
// Включает в себя функции для обработки информации о заказах пользователя
package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
	"github.com/FollowLille/loyalty/internal/utils"
)

// GetOrders возвращает информацию о заказах пользователя
//
// Параметры:
//   - c: контекст HTTP-запроса.
func GetOrders(c *gin.Context) {
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

	orders, err := database.GetUserOrders(userIDInt)
	if err != nil {
		config.Logger.Error("Failed to get orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}
	if len(orders) == 0 {
		config.Logger.Info("No orders found")
		c.JSON(http.StatusNoContent, nil)
		return
	}

	response := make([]map[string]interface{}, len(orders))
	for i, order := range orders {
		response[i] = map[string]interface{}{
			"number":      order.Number,
			"status":      order.Status,
			"accrual":     order.Accrual,
			"uploaded_at": order.UploadedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}

// UploadOrder обрабатывает информацию о заказе пользователя
//
// Параметры:
//   - c: контекст HTTP-запроса.
func UploadOrder(c *gin.Context) {
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

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	orderNumber := strings.TrimSpace(string(body))

	if !utils.CheckLunar(orderNumber) {
		config.Logger.Error("Failed to check order number", zap.Error(errors.New("invalid order number")))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		return
	}

	ownerID, err := database.GetOrderOwner(orderNumber)
	if err != nil {
		config.Logger.Error("Failed to get order owner", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order owner"})
		return
	}

	if ownerID != nil {
		if *ownerID == userIDInt {
			config.Logger.Error("Failed to create order", zap.Error(errors.New("order already uploaded by current user")))
			c.JSON(http.StatusOK, gin.H{"error": "order already uploaded by you"})
			return
		}
		config.Logger.Error("Failed to create order", zap.Error(errors.New("order already uploaded by another user")))
		c.JSON(http.StatusConflict, gin.H{"error": "order already uploaded by another user"})
		return
	}

	err = database.CreateOrder(userIDInt, orderNumber)
	if err != nil {
		config.Logger.Error("Failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}
	config.Logger.Info("Order created successfully")
	c.JSON(http.StatusAccepted, gin.H{"message": "order accepted for processing"})
}
