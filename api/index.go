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

var handler http.Handler

func init() {
	config.LoadEnv()
	db.Connect()

	if err := ai.Init(); err != nil {
		log.Printf("Failed to initialize Gemini: %v", err)
	}

	if db.DB != nil {
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
		} else {
			log.Println("Migration complete.")
		}
	} else {
		log.Println("Skipping migration because DB connection failed.")
	}

	handler = routes.SetupVercelRouter()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
