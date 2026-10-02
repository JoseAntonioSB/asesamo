// Package main is the application entrypoint for asesamo-api.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adapterHTTP "asesamo-api/internal/adapter/http"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/adapter/postgres"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/internal/infra/config"
	"asesamo-api/internal/infra/database"
)

func main() {
	log.Println("ASESAMO API starting up...")

	// 1. Load and validate environment configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Fatal configuration error: %v", err)
	}

	log.Printf("Environment: %s", cfg.AppEnv)
	log.Printf("Connecting to database at %s...", cfg.MaskedDatabaseURL())

	// 2. Initialize PostgreSQL connection pool with health check
	db, err := database.NewPostgresPool(cfg.DatabaseURL, database.DefaultPoolConfig())
	if err != nil {
		log.Fatalf("Fatal database error: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection pool: %v", err)
		}
	}()
	log.Println("Database connection pool established and verified successfully.")

	// 3. Clean Architecture Dependency Injection
	// Repositories (outbound adapters)
	userRepo := postgres.NewUserRepository(db.DB)
	accessPointRepo := postgres.NewAccessPointRepository(db.DB)

	// Use cases (core business logic)
	loginUseCase := auth.NewLoginUseCase(userRepo, cfg.JWTSecret)
	listAccessPointsUseCase := access.NewListAccessPointsUseCase(accessPointRepo)

	// Handlers (inbound adapters)
	healthHandler := handler.NewHealthHandler(db, "1.0.0")
	authHandler := handler.NewAuthHandler(loginUseCase)
	accessPointHandler := handler.NewAccessPointHandler(listAccessPointsUseCase)

	// 4. Assemble HTTP Router
	router := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:      healthHandler,
		AuthHandler:        authHandler,
		AccessPointHandler: accessPointHandler,
		JWTSecret:          cfg.JWTSecret,
		CORSOrigins:        cfg.CORSAllowedOrigins,
	})

	serverAddr := fmt.Sprintf(":%d", cfg.AppPort)
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 5. Start HTTP server in a goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	// 6. Graceful shutdown listening for SIGINT and SIGTERM
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrCh:
		log.Fatalf("Fatal HTTP server error: %v", err)
	case sig := <-quitCh:
		log.Printf("Received signal %v. Initiating graceful shutdown...", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown forced with error: %v", err)
	} else {
		log.Println("Server gracefully stopped.")
	}

	log.Println("ASESAMO API shutdown complete.")
}
