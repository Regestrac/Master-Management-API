package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Only load .env when running locally
	if os.Getenv("RENDER") == "" {
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}
}
