package routes

import (
	"fmt"
	"master-management-api/internal/handlers/auth"
	"master-management-api/internal/handlers/profile"
	"master-management-api/internal/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.POST("/signup", auth.SignUp)
	router.POST("/login", auth.Login)
	router.GET("/validate", middleware.RequireAuth, auth.Validate)
	router.POST("/logout", middleware.RequireAuth, auth.Logout)

	router.GET("/profile", middleware.RequireAuth, profile.GetProfile)

	fmt.Println("Listening to port" + os.Getenv("PORT"))
	router.Run(os.Getenv("PORT"))
}
