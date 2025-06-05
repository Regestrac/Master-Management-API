package main

import (
	"master-management-api/cmd/config"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
)

func init() {
	config.LoadEnv()
	db.Connect()
}

func main() {
	db.DB.AutoMigrate(&models.User{})
}
