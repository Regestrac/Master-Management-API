package main

import (
	"log"

	"github.com/Regestrac/Master-Management-API/cmd/config"
	"github.com/Regestrac/Master-Management-API/internal/db"
	"github.com/Regestrac/Master-Management-API/internal/models"
	"github.com/Regestrac/Master-Management-API/internal/routes"
	"github.com/Regestrac/Master-Management-API/pkg/ai"
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
