package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", time.Now())
		}
	}
}

var (
	RootPath = getRootPath()
)

func getRootPath() string {
	rootPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return rootPath
}
