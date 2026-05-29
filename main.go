package main

import (
	"log"
	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/routes"
	"master-management-api/pkg/ai"
	"net/http"
	"os"
)

func main() {
	config.LoadEnv()
	db.Connect()

	if err := ai.Init(); err != nil {
		log.Fatalf("Failed to initialize Gemini: %v", err)
	}

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

	handler := routes.SetupVercelRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
