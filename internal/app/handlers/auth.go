// Package handlers предоставляет функции для обработки запросов на взаимодействие с программой лояльности
// Включает в себя функции для обработки запросов на регистрацию и вход пользователя
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/FollowLille/loyalty/internal/config"
	cstmerr "github.com/FollowLille/loyalty/internal/errors"
	"github.com/FollowLille/loyalty/internal/services"
)

// Register обрабатывает POST-запрос на регистрацию нового пользователя.
// Если пользователь существует, возвращает ошибку cstmerr.ErrorUserAlreadyExists.
// Если пользователь не существует, создает нового пользователя и возвращает сообщение "Successful registration".
//
// Параметры:
//   - c: контекст запроса.
func Register(c *gin.Context) {
	var user struct {
		Username string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		config.Logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	token, err := services.RegisterUser(user.Username, string(hashedPassword))
	if err != nil {
		if errors.Is(err, cstmerr.ErrorUserAlreadyExists) {
			config.Logger.Warn("User already exists", zap.String("user", user.Username))
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		config.Logger.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	config.Logger.Info("User registered", zap.String("user", user.Username))

	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "Successful registration"})
}

// Login обрабатывает POST-запрос на вход в систему лояльности.
// Если пользователь существует, создает новый токен и возвращает его в качестве ответа.
//
// Параметры:
//   - c: контекст запроса.
func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		config.Logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := services.LoginUser(loginData.Username, loginData.Password)
	if err != nil {
		if errors.Is(err, cstmerr.ErrorUserDoesNotExist) || err.Error() == "invalid password" {
			config.Logger.Error("User does not exist", zap.String("user", loginData.Username))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			config.Logger.Error("Failed to process login", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "Successful login"})
}

// Logout обрабатывает POST-запрос на выход из системы лояльности.
