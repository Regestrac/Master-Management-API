package main

import (
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
	routes.SetupRouter()
	db.DB.AutoMigrate(&models.User{})
}
