package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func Connect() {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("POSTGRES_DSN"), // e.g., "host=localhost user=gorm password=gorm dbname=gorm port=9920"
		PreferSimpleProtocol: true,                      // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	log.Println("Connected to DB.")

	DB = db
}
