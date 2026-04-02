package main

import (
	"log"

	"github.com/vladislavgnilitskii/asu-soit/internal/config"
	"github.com/vladislavgnilitskii/asu-soit/internal/db"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
	"github.com/vladislavgnilitskii/asu-soit/internal/router"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	defer pool.Close()

	log.Println("подключение к БД успешно")

	// репозитории
	clientRepo  := repository.NewClientRepository(pool)
	requestRepo := repository.NewRequestRepository(pool)

	// хендлеры
	clientHandler  := handler.NewClientHandler(clientRepo)
	requestHandler := handler.NewRequestHandler(requestRepo)

	// роутер
	r := router.Setup(clientHandler, requestHandler)

	log.Printf("сервер запущен на порту %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}