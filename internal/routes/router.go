package routes

import (
	"fmt"
	"master-management-api/internal/handlers/auth"
	"master-management-api/internal/handlers/history"
	"master-management-api/internal/handlers/profile"
	"master-management-api/internal/handlers/subtasks"
	"master-management-api/internal/handlers/task"
	"master-management-api/internal/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.POST("/signup", auth.SignUp)
	router.POST("/login", auth.Login)

	router.Use(middleware.RequireAuth)

	router.GET("/validate", auth.Validate)
	router.POST("/logout", auth.Logout)

	router.GET("/profile", profile.GetProfile)
	router.PUT("/profile", profile.UpdateProfile)

	router.POST("/task", task.CreateTask)
	router.GET("/tasks", task.GetAllTasks)
	router.DELETE("/tasks/:id", task.DeleteTask)
	router.GET("/tasks/:id", task.GetTask)
	router.PATCH("/tasks/:id", task.UpdateTask)

	router.GET("/tasks/:id/history", history.GetTaskHistory)
	router.POST("/tasks/:id/history", history.AddToHistory)
	router.POST("/task/:id/generate-description", task.GenerateDescription)

	router.GET("/tasks/:id/subtasks", subtasks.GetAllSubtasks)

	router.Run(os.Getenv("PORT"))
	fmt.Println("Listening to port" + os.Getenv("PORT"))
}
