package agent

import (
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
)

type OrderAgent struct {
	Interval time.Duration
}

var statuses = []string{"PROCESSING", "PROCESSED", "INVALID"}

func generateRandomStatus() string {
	return statuses[time.Now().UnixNano()%3]
}

func generateRandomAccrual() int64 {
	return time.Now().UnixNano() % 1000
}

func generateStatusAndAccrual() (string, int64) {
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
			status, accrual := generateStatusAndAccrual()
			err = database.UpdateOrder(order.Number, status, accrual)
			if err != nil {
				config.Logger.Error("Failed to update order", zap.Error(err))
				time.Sleep(a.Interval)
				continue
			}

			config.Logger.Info("Updated order",
				zap.String("order_number", order.Number),
				zap.String("status", status),
				zap.Int64("accrual", accrual),
			)

		}
		time.Sleep(a.Interval)
	}
}

func StartAgent() {
	agent := &OrderAgent{
		Interval: 5 * time.Second,
	}
	go agent.processOrders()
}
