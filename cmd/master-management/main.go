package main

import (
	"log"

	"github.com/Regestrac/master-management-api/cmd/config"
	"github.com/Regestrac/master-management-api/internal/db"
	"github.com/Regestrac/master-management-api/internal/models"
	"github.com/Regestrac/master-management-api/internal/routes"
	"github.com/Regestrac/master-management-api/pkg/ai"
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
		&models.UserSettings{},
		&models.TaskSession{},
	); err != nil {
		log.Fatal("Failed to migrate:", err)
	}
	log.Println("Migration complete.")

	routes.SetupRouter()
}
