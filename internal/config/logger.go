// Package config хранит конфигурацию приложения.
// Включает функции для инициализации логгера и подключения к базе данных.
package config

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger = zap.NewNop()

// InitLogger инициализирует логгер.
// Если произошла ошибка при инициализации, программа завершается с кодом ошибки.
//
// Параметры:
//   - level: уровень логирования.
//
// Возвращает:
//   - error: ошибка, если произошла ошибка при инициализации логгера.
func InitLogger(level string) error {
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(logLevel)

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}
	return nil
}

// RequestLogger инициализирует обработчик для логирования запросов.
// В лог попадают все входящие запросы и ответы.
//
// Возвращает:
//   - gin.HandlerFunc: обработчик для логирования запросов.
//
// Логирует:
//   - метод запроса
//   - путь запроса
//   - время выполнения запроса
//   - тело запроса
//   - хэдеры запроса
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			Logger.Error("Failed to read request body", zap.Error(err))
			c.Next()
			return
		}

		headers := c.Request.Header
		headerMap := make(map[string]string)
		for k, v := range headers {
			headerMap[k] = v[0]
		}

		Logger.Info("Got incomming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.ByteString("body", bodyBytes),
			zap.Any("headers", headerMap),
		)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Set("requestBody", bodyBytes)
		c.Next()
	}
}

// ResponseLogger инициализирует обработчик для логирования ответов.
// В лог попадает только ответ.
//
// Возвращает:
//   - gin.HandlerFunc: обработчик для логирования ответов.
//
// Логирует:
//   - статус код ответа
//   - размер ответа
//   - тело ответа
func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		if requestBody, exists := c.Get("requestBody"); exists {
			Logger.Info("Sent response",
				zap.Int("status_code", statusCode),
				zap.Int("response_size", responseSize),
				zap.ByteString("request_body", requestBody.([]byte)),
			)
		} else {
			Logger.Warn("Request body not found in context")
		}
	}
}
