package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Небезопасные значения по умолчанию — годятся только для локальной
// разработки. В production они запрещены (см. проверку в Load).
const (
	defaultDBPassword = "localdevpassword"
	defaultJWTSecret  = "dev-secret-change-in-production"
)

type Config struct {
	AppEnv     string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	AppPort    string
	JWTSecret  string
}

func Load() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("файл .env не найден, читаем переменные окружения")
	}

	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "asu_soit"),
		DBUser:     getEnv("DB_USER", "asu_soit_user"),
		DBPassword: getEnv("DB_PASSWORD", defaultDBPassword),
		AppPort:    getEnv("APP_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", defaultJWTSecret),
	}

	cfg.mustBeSecureInProduction()
	return cfg
}

// mustBeSecureInProduction не даёт запустить сервис в production
// с дефолтными секретами — это фатальная ошибка конфигурации.
func (c *Config) mustBeSecureInProduction() {
	if c.AppEnv != "production" {
		return
	}
	if c.JWTSecret == "" || c.JWTSecret == defaultJWTSecret {
		log.Fatal("APP_ENV=production, но JWT_SECRET не задан или равен дефолтному — задайте безопасный секрет")
	}
	if c.DBPassword == "" || c.DBPassword == defaultDBPassword {
		log.Fatal("APP_ENV=production, но DB_PASSWORD не задан или равен дефолтному — задайте безопасный пароль")
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}
