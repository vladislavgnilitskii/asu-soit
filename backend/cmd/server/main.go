package main

import (
	"log"

	"github.com/vladislavgnilitskii/asu-soit/internal/config"
	"github.com/vladislavgnilitskii/asu-soit/internal/db"
)

func main() {
	cfg := config.Load()
	log.Printf("конфиг загружен: хост БД = %s, порт приложения = %s",
		cfg.DBHost, cfg.AppPort)

	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	defer pool.Close()

	log.Println("подключение к БД успешно")
	log.Println("сервер готов к запуску")
}
