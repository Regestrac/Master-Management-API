package handler

import (
	"log"
	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/routes"
	"master-management-api/pkg/ai"
	"net/http"
)

var handler http.Handler

func init() {
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

	handler = routes.SetupVercelRouter()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
