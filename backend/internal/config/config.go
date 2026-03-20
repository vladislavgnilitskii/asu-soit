package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	AppPort    string
}

func Load() *Config {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("файл .env не найден, читаем переменные окружения напрямую")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "asu_soit"),
		DBUser:     getEnv("DB_USER", "asu_soit_user"),
		DBPassword: getEnv("DB_PASSWORD", "localdevpassword"),
		AppPort:    getEnv("APP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}
