package routes

import (
	"fmt"
	"master-management-api/internal/handlers/auth"
	"master-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.POST("/signup", auth.SignUp)
	router.POST("/login", auth.Login)
	router.GET("/validate", middleware.RequireAuth, auth.Validate)

	fmt.Println("Listening to port 8080")
	router.Run(":8080")
}
