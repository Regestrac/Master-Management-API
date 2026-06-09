package api

import (
	"log"
	"net/http"

	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/routes"
	"master-management-api/pkg/ai"
)

var app http.Handler

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
		log.Printf("Failed to migrate: %v", err)
	}
	log.Println("Migration complete.")

	app = routes.SetupVercelRouter()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
