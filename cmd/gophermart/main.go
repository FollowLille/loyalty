// Package main отвечает за инициализацию и запуск сервера лояльности.
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/FollowLille/loyalty/internal/agent"
	"github.com/FollowLille/loyalty/internal/app/handlers"
	"github.com/FollowLille/loyalty/internal/app/middleware"
	"github.com/FollowLille/loyalty/internal/compress"
	"github.com/FollowLille/loyalty/internal/config"
	"github.com/FollowLille/loyalty/internal/database"
)

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		fmt.Println("Failed to load .env file")
	}

	parseFlags()

	if err := config.InitLogger(flagLogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %s\n", err.Error())
		os.Exit(1)
	}
	config.Logger.Info("Logger initialized")

	if err := prepareDB(); err != nil {
		config.Logger.Error("Failed to prepare database", zap.Error(err))
		os.Exit(1)
	}

	router := gin.New()
	router.Use(gin.Recovery(), config.RequestLogger(), config.ResponseLogger())
	router.Use(compress.GzipMiddleware(), compress.GzipResponseMiddleware())

	public := router.Group("/api/user")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
	}

	protected := router.Group("/api/user")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/orders", handlers.UploadOrder)
		protected.GET("/orders", handlers.GetOrders)
		protected.GET("/balance", handlers.GetBalance)
		protected.POST("/balance/withdraw", handlers.GetWithdrawRequest)
		protected.GET("/withdrawals", handlers.GetWithdrawals)
	}

	config.Logger.Info("Starting server...", zap.String("address", flagAddress))

	agent.StartAgent(flagAPI)
	router.Run(flagAddress)
}

func prepareDB() error {
	config.Logger.Info("Preparing database")
	if err := database.InitDB(flagDatabaseAddress); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	if err := database.PrepareDB(); err != nil {
		return fmt.Errorf("failed to prepare database: %w", err)
	}
	config.Logger.Info("Database prepared")
	return nil
}
