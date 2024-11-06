// Package handlers предоставляет функции для обработки запросов на взаимодействие с программой лояльности
package handlers

import (
	"errors"
	"github.com/FollowLille/loyalty/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/FollowLille/loyalty/internal/database"
	cstmerr "github.com/FollowLille/loyalty/internal/errors"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	err = database.CreateUser(user.Username, string(hashedPassword))
	if err != nil {
		if errors.Is(err, cstmerr.ErrorUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successful registration"})
}

// Login обрабатывает POST-запрос на вход в систему лояльности.
// Если пользователь не существует, возвращает ошибку cstmerr.ErrorUserDoesNotExist.
// Если пользователь существует, проверяет пароль.
// Если пароль некорректный, возвращает ошибку cstmerr.ErrorInvalidPassword.
// Если пароль корректный, то возвращает статус-код 200 и сообщение "Successful login".
//
// Параметры:
//   - c: контекст запроса.
func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := database.GetUserPasswordHash(loginData.Username)
	if err != nil {
		if errors.Is(err, cstmerr.ErrorUserDoesNotExist) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user), []byte(loginData.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateToken(loginData.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successful login",
		"token":   token,
	})
}
