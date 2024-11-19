package mock

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/config"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var statuses = []string{"PROCESSING", "PROCESSED", "INVALID"}

func StartMockAccrualServer(address string) {
	http.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		orderNumber := r.URL.Path[len("/api/orders/"):]

		if orderNumber == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		config.Logger.Info("Get order from external API", zap.String("order_number", orderNumber))

		status := statuses[rand.Intn(len(statuses))]
		var accrual float64
		if status == "PROCESSED" {
			accrual = float64(rand.Intn(1000000)) / 100
		}

		response := AccrualResponse{
			Order:   orderNumber,
			Status:  status,
			Accrual: accrual,
		}

		// Возвращаем ответ в формате JSON.
		w.Header().Set("Content-Type", "application/json")
		if status == "INVALID" || status == "PROCESSING" || status == "REGISTERED" {
			response.Accrual = 0 // Не показываем начисления для неокончательных статусов.
		}

		switch status {
		case "REGISTERED", "INVALID", "PROCESSING", "PROCESSED":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNoContent) // Если статус "заказ не найден".
		}
	})

	// Запускаем сервер.
	config.Logger.Info("Starting mock accrual server", zap.String("address", address))
	if err := http.ListenAndServe(address, nil); err != nil {
		config.Logger.Fatal("Failed to start mock accrual server", zap.Error(err))
	}
}
