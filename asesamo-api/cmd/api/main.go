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
	"asesamo-api/internal/adapter/storage"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/internal/core/usecase/document"
	"asesamo-api/internal/core/usecase/user"
	"asesamo-api/internal/core/usecase/worker"
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

	// Repositories (outbound adapters)
	userRepo := postgres.NewUserRepository(db.DB)
	accessPointRepo := postgres.NewAccessPointRepository(db.DB)
	plantRepo := postgres.NewPlantRepository(db.DB)
	accessLogRepo := postgres.NewAccessLogRepository(db.DB)
	complianceRepo := postgres.NewComplianceRepository(db.DB)
	docRepo := postgres.NewDocumentRepository(db.DB)

	storageService, err := storage.NewLocalStorageService(cfg.StorageDir)
	if err != nil {
		log.Fatalf("Fatal storage error: %v", err)
	}

	// Use cases (core business logic)
	loginUseCase := auth.NewLoginUseCase(userRepo, cfg.JWTSecret)
	listAccessPointsUseCase := access.NewListAccessPointsUseCase(accessPointRepo)
	checkInUseCase := access.NewCheckInUseCase(accessPointRepo, plantRepo, complianceRepo, accessLogRepo)
	checkOutUseCase := access.NewCheckOutUseCase(accessLogRepo, accessPointRepo)
	listAccessLogsUseCase := access.NewListAccessLogsUseCase(accessLogRepo)

	uploadDocUseCase := document.NewUploadDocumentUseCase(docRepo, storageService)
	listMyDocsUseCase := document.NewListMyDocumentsUseCase(docRepo)
	downloadDocUseCase := document.NewDownloadDocumentUseCase(docRepo, storageService)
	listDocTypesUseCase := document.NewListDocumentTypesUseCase(docRepo)
	reviewDocUseCase := document.NewReviewDocumentUseCase(docRepo)
	listPendingDocsUseCase := document.NewListPendingDocumentsUseCase(docRepo)
	updateThemeUseCase := user.NewUpdateThemeUseCase(userRepo)
	getCredentialUseCase := worker.NewGetCredentialUseCase(userRepo, plantRepo, complianceRepo)

	// Handlers (inbound adapters)
	healthHandler := handler.NewHealthHandler(db, "1.0.0")
	authHandler := handler.NewAuthHandler(loginUseCase)
	accessPointHandler := handler.NewAccessPointHandler(listAccessPointsUseCase)
	accessHandler := handler.NewAccessHandler(checkInUseCase, checkOutUseCase, listAccessLogsUseCase)
	docHandler := handler.NewDocumentHandler(uploadDocUseCase, listMyDocsUseCase, downloadDocUseCase, listDocTypesUseCase)
	docHandler.SetReviewUseCases(reviewDocUseCase, listPendingDocsUseCase)
	userHandler := handler.NewUserHandler(updateThemeUseCase)
	workerHandler := handler.NewWorkerHandler(getCredentialUseCase)

	loginLimit := 100
	if cfg.AppEnv == config.EnvProduction {
		loginLimit = 10
	}

	// 4. Assemble HTTP Router
	router := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:      healthHandler,
		AuthHandler:        authHandler,
		AccessPointHandler: accessPointHandler,
		AccessHandler:      accessHandler,
		DocumentHandler:    docHandler,
		UserHandler:        userHandler,
		WorkerHandler:      workerHandler,
		JWTSecret:          cfg.JWTSecret,
		CORSOrigins:        cfg.CORSAllowedOrigins,
		LoginRateLimit:     loginLimit,
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
