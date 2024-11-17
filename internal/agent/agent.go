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

func generateRandomStatus() string {
	return statuses[time.Now().UnixNano()%3]
}

func generateRandomAccrual() float64 {
	return float64(time.Now().UnixNano() % 1000)
}

func generateStatusAndAccrual() (string, float64) {
	status := generateRandomStatus()
	if status == "PROCESSED" {
		return status, generateRandomAccrual()
	}
	return status, 0
}

func (a *OrderAgent) processOrders() {
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

func StartAgent() {
	agent := &OrderAgent{
		Interval:       5 * time.Second,
		UseExternalAPI: false,
	}
	go agent.processOrders()
}

func StartAgentExternalApi() {
	agent := &OrderAgent{
		Interval:       5 * time.Second,
		UseExternalAPI: true,
	}
	go agent.processOrders()
}
