package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
	cstmerr "github.com/FollowLille/loyalty/internal/errors"
)

// ExternalAccrualResponse описывает структуру ответа от внешней системы начислений.
type ExternalAccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func FetchOrderAccrual(orderNumber string) (*ExternalAccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", config.AccrualAPIURL, orderNumber)
	config.Logger.Info("Requesting external accrual API", zap.String("url", url))

	client := &http.Client{
		Timeout: 30 * time.Second, // Устанавливаем таймаут на запрос.
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		config.Logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Error("Failed to perform request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK: // 200 OK
		var response ExternalAccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			config.Logger.Error("Failed to decode response", zap.Error(err))
			return nil, err
		}
		return &response, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter == "" {
			retryAfter = "60"
		}
		config.Logger.Warn("Too many requests", zap.String("retry-after", retryAfter))
		return nil, fmt.Errorf("too many requests, retry after %s seconds", retryAfter)

	case http.StatusNoContent: // 204 No Content
		config.Logger.Info("Order not found in external system", zap.String("order", orderNumber))
		return nil, cstmerr.ErrOrderNotFound

	case http.StatusInternalServerError: // 500 Internal Server Error
		config.Logger.Error("External API returned internal server error")
		return nil, fmt.Errorf("internal server error from external API")

	default:
		config.Logger.Error("Unexpected status code from external API", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func ProcessOrderAccrual(orderNumber string) error {
	config.Logger.Info("Processing order accrual", zap.String("order", orderNumber))

	response, err := FetchOrderAccrual(orderNumber)
	if err != nil {
		return err
	}

	err = database.UpdateOrder(orderNumber, response.Status, response.Accrual)
	if err != nil {
		config.Logger.Error("Failed to update order", zap.Error(err))
		return fmt.Errorf("failed to update order: %w", err)
	}

	config.Logger.Info("Order processed", zap.String("order", orderNumber))
	return nil
}
