package main

import (
	"github.com/FollowLille/loyalty/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api/user")
	{
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
	}

	router.Run(":8080")
}
