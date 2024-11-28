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
	stopCh         chan struct{} // канал для завершения работы агента
}

var statuses = []string{"PROCESSING", "PROCESSED", "INVALID"}

// generateRandomOrder генерирует случайный статус
func generateRandomStatus() string {
	return statuses[time.Now().UnixNano()%3]
}

// generateRandomAccrual генерирует случайные начисления
func generateRandomAccrual() float64 {
	//return float64(time.Now().UnixNano() % 1000)
	mockFloat := 729.98
	return mockFloat
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
// Агент постоянно ходит в базу данных, проверяет наличие необработанных заказов и обрабатывает их
func (a *OrderAgent) processOrders() {
	config.Logger.Info("Process orders started")
	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			orders, err := database.GetOrdersByStatus()
			if err != nil {
				config.Logger.Error("Failed to get orders", zap.Error(err))
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

				err = database.UpdateOrder(order.Number, status, accrual)
				if err != nil {
					config.Logger.Error("Failed to update order", zap.Error(err))
					continue
				}

				config.Logger.Info("Updated order",
					zap.String("order_number", order.Number),
					zap.String("status", status),
					zap.Float64("accrual", accrual),
				)
			}
		case <-a.stopCh:
			config.Logger.Info("Stopping order processing")
			return
		}
	}
}

// StartAgent запускает агента с генерацией случайных данных
func StartAgent(apiFlag bool) *OrderAgent {
	agent := &OrderAgent{
		Interval:       5 * time.Second,
		UseExternalAPI: apiFlag,
		stopCh:         make(chan struct{}),
	}
	go agent.processOrders()
	return agent
}

// StopAgent завершает работу агента
func (a *OrderAgent) StopAgent() {
	close(a.stopCh)
	config.Logger.Info("Agent stopped")
}
