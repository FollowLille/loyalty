package agent

import (
	"time"

	"go.uber.org/zap"

	accrualHandler "github.com/FollowLille/loyalty/internal/accrual"
	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
)

type OrderAgent struct {
	Interval       time.Duration
	UseExternalAPI bool
}

var statuses = []string{"PROCESSING", "PROCESSED", "INVALID"}

// generateRandomOrder генерирует случайный статус
func generateRandomStatus() string {
	return statuses[time.Now().UnixNano()%3]
}

// generateRandomAccrual генерирует случайные начисления
func generateRandomAccrual() float64 {
	return float64(time.Now().UnixNano() % 1000)
}

// generateStatusAndAccrual генерирует случайные статус и начисления
func generateStatusAndAccrual() (string, float64) {
	status := generateRandomStatus()
	if status == "PROCESSED" {
		return status, generateRandomAccrual()
	}
	return status, 0
}

// processOrders обрабатывает актуальные заказы
// Агент постоянно ходит в базу данных, проверяет наличие необработанных заказов и обрбатывает их
func (a *OrderAgent) processOrders() {
	time.Sleep(5 * time.Second)
	for {
		orders, err := database.GetOrdersByStatus()
		if err != nil {
			config.Logger.Error("Failed to get orders", zap.Error(err))
			time.Sleep(a.Interval)
			continue
		}
		for _, order := range orders {
			var status string
			var accrual float64

			if a.UseExternalAPI {
				config.Logger.Info("Get order from external API", zap.String("order_number", order.Number))
				response, err := accrualHandler.FetchOrderAccrual(order.Number)
				if err != nil {
					config.Logger.Error("Failed to get order from external API", zap.Error(err))
					continue
				}

				status = response.Status
				accrual = response.Accrual
			} else {
				status, accrual = generateStatusAndAccrual()
			}
			err = database.UpdateOrder(order.Number, status, float64(accrual))
			if err != nil {
				config.Logger.Error("Failed to update order", zap.Error(err))
				time.Sleep(a.Interval)
				continue
			}

			config.Logger.Info("Updated order",
				zap.String("order_number", order.Number),
				zap.String("status", status),
				zap.Float64("accrual", accrual),
			)

		}
		time.Sleep(a.Interval)
	}
}

// StartAgent запускает агента с генерацией случайных данных
func StartAgent() {
	agent := &OrderAgent{
		Interval:       5 * time.Second,
		UseExternalAPI: false,
	}
	go agent.processOrders()
}

// StartAgentExternalAPI запускает агента с использованием внешнего API
func StartAgentExternalAPI() {
	agent := &OrderAgent{
		Interval:       5 * time.Second,
		UseExternalAPI: true,
	}
	go agent.processOrders()
}
