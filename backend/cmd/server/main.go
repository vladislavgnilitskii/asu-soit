package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	clientRepo := repository.NewClientRepository(pool)
	requestRepo := repository.NewRequestRepository(pool)
	deviceRepo := repository.NewDeviceRepository(pool)
	warehouseRepo := repository.NewWarehouseRepository(pool)
	invoiceRepo := repository.NewInvoiceRepository(pool)
	employeeRepo := repository.NewEmployeeRepository(pool)

	// хендлеры
	clientHandler := handler.NewClientHandler(clientRepo)
	requestHandler := handler.NewRequestHandler(requestRepo)
	deviceHandler := handler.NewDeviceHandler(deviceRepo)
	warehouseHandler := handler.NewWarehouseHandler(warehouseRepo)
	invoiceHandler := handler.NewInvoiceHandler(invoiceRepo)
	employeeHandler := handler.NewEmployeeHandler(employeeRepo)
	authHandler := handler.NewAuthHandler(employeeRepo, cfg.JWTSecret)

	// роутер
	r := router.Setup(clientHandler, requestHandler, deviceHandler, warehouseHandler, invoiceHandler, employeeHandler, authHandler, cfg.JWTSecret, pool)

	// таймауты защищают от медленных клиентов (Slowloris) и висящих соединений
	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// сервер поднимаем в отдельной горутине, чтобы main мог ждать сигнал
	go func() {
		log.Printf("сервер запущен на порту %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ошибка запуска сервера: %v", err)
		}
	}()

	// graceful shutdown: по SIGINT/SIGTERM даём активным запросам завершиться
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("останавливаем сервер…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("некорректная остановка сервера: %v", err)
	}
	log.Println("сервер остановлен")
}
