package routes

import (
	"log"
	"master-management-api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() {
	router := gin.Default()

	router.GET("/tasks", handlers.GetTasks)

	log.Fatal(router.Run(":8080"))
}
