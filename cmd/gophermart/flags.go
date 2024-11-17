// Package main отвечает за инициализацию и запуск сервера лояльности.
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"fmt"
	"github.com/FollowLille/loyalty/internal/config"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	flagAddress         string // Server address
	flagDatabaseAddress string // Database address
	flagAccrualAddress  string // Accrual system address
	flagLogLevel        string // Log level
)

// parseFlags парсит командные флаги и переменные окружения для настройки сервера.
// Флаги включают адрес сервера, адрес базы данных, адрес системы начисления и уровень логирования.
// Если переменные окружения определены, они имеют приоритет над значениями по умолчанию.
//
// Пример использования:
//
//	-address=127.0.0.1:8080
//	-database=postgres://user:password@localhost/dbname
//	-accrual-address=http://localhost:8081
//	-log-level=debug
//
// После парсинга флагов, информация о них логируется с использованием zap.
func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "127.0.0.1:8080", "Server address")
	pflag.StringVarP(&flagDatabaseAddress, "database", "d", "", "Database address")
	pflag.StringVarP(&flagAccrualAddress, "accrual-address", "r", "127.0.0.1:8081", "Accrual system address")
	pflag.StringVarP(&flagLogLevel, "log-level", "l", "info", "Log level")
	pflag.Parse()

	// Переопределение значений флагов значениями переменных окружения, если они заданы.
	if envAddress := os.Getenv("RUN_ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}

	if envDatabase := os.Getenv("DATABASE_URI"); envDatabase != "" {
		flagDatabaseAddress = envDatabase
	}

	if envAccuralAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccuralAddress != "" {
		flagAccrualAddress = envAccuralAddress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
	fmt.Println("Flag is: ", flagDatabaseAddress, " ", flagAccrualAddress, " ", flagLogLevel, " ", flagAddress, " ", flagDatabaseAddress)
	// Логируем значения флагов.
	config.Logger.Info("Flags parsed",
		zap.String("address", flagAddress),
		zap.String("database", flagDatabaseAddress),
		zap.String("accrual", flagAccrualAddress),
		zap.String("log-level", flagLogLevel))
	fmt.Println("Flag is: ", flagDatabaseAddress)
}
