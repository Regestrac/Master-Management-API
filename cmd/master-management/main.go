package main

import (
	"log"
	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/routes"
	"master-management-api/pkg/ai"
)

func init() {
	config.LoadEnv()
	db.Connect()

	if err := ai.Init(); err != nil {
		log.Fatalf("Failed to initialize Gemini: %v", err)
	}
}

func main() {
	log.Println("Starting Migration...")
	if err := db.DB.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.TaskHistory{},
		&models.Note{},
		&models.Checklist{},
		&models.Workspace{},
		&models.Member{},
	); err != nil {
		log.Fatal("Failed to migrate:", err)
	}
	log.Println("Migration complete.")

	routes.SetupRouter()
}
