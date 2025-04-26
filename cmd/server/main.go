package main

import (
	"AuthService/config"
	"AuthService/internal/security"
	"AuthService/internal/logger"
	_ "AuthService/internal/models"
	"AuthService/internal/domain"
	"AuthService/internal/repository"
	"AuthService/internal/service"
	"AuthService/internal/transport/http/api"

	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()

	// Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx, err = logger.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Error create logger: %v", err)
	}

	//Подключаемся к БД
	dbPool, err := domain.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	//3. Запускаем миграции
	err = domain.RunMigrations(ctx, cfg, dbPool)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Создаем репозитории
	userRepo := repository.NewRepository(dbPool, cfg)

	// Создаем security
	securityStuff := security.NewSecurity(cfg)

	//Создаем сервисы
	userService := service.NewAuthService(userRepo, securityStuff, cfg)

	// Создаем хэндлер
	handler := api.NewHandler(ctx, userService)

	// Создаём роутер
	router := api.NewRouter(handler)

	//Запускаем сервер
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s...", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Завршаем работу сервер (Graceful Shutdown)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Server gracefull stopped")
}
