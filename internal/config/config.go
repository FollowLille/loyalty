// Package config хранит конфигурацию приложения.
// Включает функции для инициализации логгера и подключения к базе данных.
package config

import "time"

// DatabaseRetryDelays хранит задержки между повторными попытками подключения к базе данных.
// Первая задержка - 1 секунда, вторая - 3 секунды, третья - 5 секунд.
var DatabaseRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

var SuperSecretKey string = "You'llNeverGuessIt"
