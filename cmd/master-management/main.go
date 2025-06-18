package main

import (
	"log"
	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/routes"
)

func init() {
	config.LoadEnv()
	db.Connect()
}

func main() {
	log.Println("Starting Migration...")
	if err := db.DB.AutoMigrate(&models.User{}, &models.Task{}, &models.TaskHistory{}); err != nil {
		log.Fatal("Failed to migrate:", err)
	}
	log.Println("Migration complete.")

	routes.SetupRouter()
}
