package api

import (
	"log"
	"os"

	"siot/api/controllers"
	"siot/api/seed"

	"github.com/joho/godotenv"
)

var server = controllers.Server{}

func init() {

	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print(".env file found")
	}
}

func Run() {

	var err error = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, %v", err)
	}

	server.Initialize(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"), os.Getenv("MONGO_HOST"))

	seed.Load(server.DB)

	server.Run(":8080")

}
