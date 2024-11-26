// Package handlers предоставляет функции для обработки HTTP-запросов в системе лояльности.
// Включает в себя функции для обработки информации о заказах пользователя
package handlers

import (
	"github.com/FollowLille/loyalty/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/FollowLille/loyalty/internal/config"
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

	orders, err := services.GetOrders(userIDInt)
	if err != nil {
		config.Logger.Error("Failed to get orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}
	if orders == nil {
		config.Logger.Info("No orders found")
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, orders)
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

	if err := services.UploadOrder(userIDInt, orderNumber); err != nil {
		config.Logger.Error("Failed to upload order", zap.Error(err))
		switch err.Error() {
		case "invalid order number":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		case "order already uploaded by current user":
			c.JSON(http.StatusOK, gin.H{"error": "order already uploaded by you"})
		case "order already uploaded by another user":
			c.JSON(http.StatusConflict, gin.H{"error": "order already uploaded by another user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload order"})
		}
		return
	}

	config.Logger.Info("Order created successfully")
	c.JSON(http.StatusAccepted, gin.H{"message": "order accepted for processing"})
}
