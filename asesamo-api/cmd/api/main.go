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
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/internal/core/usecase/contractor"
	"asesamo-api/internal/core/usecase/document"
	"asesamo-api/internal/core/usecase/epp"
	"asesamo-api/internal/core/usecase/evacuation"
	"asesamo-api/internal/core/usecase/health"
	"asesamo-api/internal/core/usecase/incident"
	"asesamo-api/internal/core/usecase/inspection"
	"asesamo-api/internal/core/usecase/permit"
	"asesamo-api/internal/core/usecase/report"
	"asesamo-api/internal/core/usecase/training"
	"asesamo-api/internal/core/usecase/user"
	"asesamo-api/internal/core/usecase/worker"
	"asesamo-api/internal/infra/config"
	"asesamo-api/internal/infra/database"
	"asesamo-api/internal/infra/notification"
	"asesamo-api/internal/infra/realtime"
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
	contractorRepo := postgres.NewContractorRepository(db.DB)
	reportRepo := postgres.NewReportRepository(db.DB)
	permitRepo := postgres.NewWorkPermitRepository(db.DB)
	inspectionRepo := postgres.NewSafetyInspectionRepository(db.DB)
	incidentRepo := postgres.NewSafetyIncidentRepository(db.DB)
	eppRepo := postgres.NewEPPRepository(db.DB)
	evacuationRepo := postgres.NewEmergencyEvacuationRepository(db.DB)
	trainingRepo := postgres.NewTrainingRepository(db.DB)
	healthRepo := postgres.NewHealthRepository(db.DB)

	complianceEngine := service.NewComplianceEngine(complianceRepo)

	var storageService repository.StorageService
	if cfg.StorageDriver == "r2" {
		log.Printf("INFO: Initializing Cloudflare R2 Object Storage (Bucket: %s, Account: %s)", cfg.R2BucketName, cfg.R2AccountID)
		r2Svc, err := storage.NewR2StorageService(storage.R2Config{
			AccountID:       cfg.R2AccountID,
			AccessKeyID:     cfg.R2AccessKeyID,
			SecretAccessKey: cfg.R2SecretAccessKey,
			BucketName:      cfg.R2BucketName,
			Endpoint:        cfg.R2Endpoint,
			PublicURL:       cfg.R2PublicURL,
		})
		if err != nil {
			log.Fatalf("Fatal R2 storage error: %v", err)
		}
		storageService = r2Svc
	} else {
		log.Printf("INFO: Initializing Local Disk Storage Adapter (Dir: %s)", cfg.StorageDir)
		localSvc, err := storage.NewLocalStorageService(cfg.StorageDir)
		if err != nil {
			log.Fatalf("Fatal local storage error: %v", err)
		}
		storageService = localSvc
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
	getProfileUseCase := user.NewGetProfileUseCase(userRepo, contractorRepo)
	updateProfileUseCase := user.NewUpdateProfileUseCase(userRepo)
	uploadAvatarUseCase := user.NewUploadAvatarUseCase(userRepo, storageService)
	getAvatarUseCase := user.NewGetAvatarUseCase(userRepo, storageService)
	getCredentialUseCase := worker.NewGetCredentialUseCase(userRepo, plantRepo, complianceRepo)

	// Contractor Use Cases
	inviteContractorUseCase := contractor.NewInviteContractorUseCase(contractorRepo, userRepo, plantRepo)
	listContractorsUseCase := contractor.NewListPlantContractorsUseCase(contractorRepo, plantRepo)
	updateContractorStatusUseCase := contractor.NewUpdateContractorStatusUseCase(contractorRepo)
	getMyCompanyUseCase := contractor.NewGetMyContractorCompanyUseCase(contractorRepo)
	listWorkersUseCase := contractor.NewListContractorWorkersUseCase(contractorRepo)
	addWorkerUseCase := contractor.NewAddContractorWorkerUseCase(contractorRepo, userRepo)
	toggleWorkerStatusUseCase := contractor.NewToggleContractorWorkerStatusUseCase(contractorRepo)

	// Report Use Cases (Phase 2 Module 2)
	getMetricsUseCase := report.NewGetReportMetricsUseCase(reportRepo)
	exportAccessLogsUseCase := report.NewExportAccessLogsCSVUseCase(reportRepo)
	exportComplianceUseCase := report.NewExportComplianceCSVUseCase(reportRepo)

	// PTAR Digital Use Cases (Milestone 2)
	createPermitUseCase := permit.NewCreatePermitUseCase(permitRepo)
	listPermitsUseCase := permit.NewListPermitsUseCase(permitRepo, contractorRepo)
	getPermitUseCase := permit.NewGetPermitUseCase(permitRepo, contractorRepo)
	assignWorkersUseCase := permit.NewAssignWorkersUseCase(permitRepo, complianceEngine)
	updateChecklistUseCase := permit.NewUpdateChecklistUseCase(permitRepo)
	approvePermitUseCase := permit.NewApprovePermitUseCase(permitRepo, complianceEngine)
	transitionStatusUseCase := permit.NewTransitionStatusUseCase(permitRepo)

	// Operational Safety Use Cases (Milestone 3)
	completeInspectionUseCase := inspection.NewCompleteInspectionUseCase(inspectionRepo)
	reportIncidentUseCase := incident.NewReportIncidentUseCase(incidentRepo)
	deliverEPPUseCase := epp.NewDeliverEPPUseCase(eppRepo)
	startEvacuationUseCase := evacuation.NewStartEvacuationUseCase(evacuationRepo)

	// Training & Health Use Cases (Phase 4)
	issueDC3UseCase := training.NewIssueDC3CertificateUseCase(trainingRepo)
	recordMedicalExamUseCase := health.NewRecordMedicalExamUseCase(healthRepo)

	// Handlers (inbound adapters)
	healthHandler := handler.NewHealthHandler(db, "1.0.0")
	authHandler := handler.NewAuthHandler(loginUseCase)
	accessPointHandler := handler.NewAccessPointHandler(listAccessPointsUseCase)
	accessHandler := handler.NewAccessHandler(checkInUseCase, checkOutUseCase, listAccessLogsUseCase)
	docHandler := handler.NewDocumentHandler(uploadDocUseCase, listMyDocsUseCase, downloadDocUseCase, listDocTypesUseCase)
	docHandler.SetReviewUseCases(reviewDocUseCase, listPendingDocsUseCase)
	notificationService := notification.NewService(
		notification.NewMockEmailSender(),
		notification.NewHTTPWebhookSender(5*time.Second),
		"",
		"",
	)
	docHandler.SetNotificationDispatcher(notificationService)
	userHandler := handler.NewUserHandler(
		getProfileUseCase,
		updateProfileUseCase,
		uploadAvatarUseCase,
		getAvatarUseCase,
		updateThemeUseCase,
	)
	workerHandler := handler.NewWorkerHandler(getCredentialUseCase)
	contractorHandler := handler.NewContractorHandler(
		inviteContractorUseCase,
		listContractorsUseCase,
		updateContractorStatusUseCase,
		getMyCompanyUseCase,
		listWorkersUseCase,
		addWorkerUseCase,
		toggleWorkerStatusUseCase,
	)
	reportHandler := handler.NewReportHandler(getMetricsUseCase, exportAccessLogsUseCase, exportComplianceUseCase)
	permitHandler := handler.NewPermitHandler(
		createPermitUseCase,
		listPermitsUseCase,
		getPermitUseCase,
		assignWorkersUseCase,
		updateChecklistUseCase,
		approvePermitUseCase,
		transitionStatusUseCase,
	)
	inspectionHandler := handler.NewInspectionHandler(inspectionRepo, completeInspectionUseCase)
	incidentHandler := handler.NewIncidentHandler(incidentRepo, reportIncidentUseCase)
	eppHandler := handler.NewEPPHandler(eppRepo, deliverEPPUseCase)
	eventBroker := realtime.NewInMemoryEventBroker(100)
	realtimeHandler := handler.NewRealtimeHandler(eventBroker)
	evacuationHandler := handler.NewEvacuationHandler(evacuationRepo, startEvacuationUseCase, eventBroker)
	trainingHandler := handler.NewTrainingHandler(trainingRepo, issueDC3UseCase)
	occupationalHealthHandler := handler.NewOccupationalHealthHandler(healthRepo, recordMedicalExamUseCase)

	loginLimit := 100
	if cfg.AppEnv == config.EnvProduction {
		loginLimit = 10
	}

	// 4. Assemble HTTP Router
	router := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:             healthHandler,
		AuthHandler:               authHandler,
		AccessPointHandler:        accessPointHandler,
		AccessHandler:             accessHandler,
		DocumentHandler:           docHandler,
		UserHandler:               userHandler,
		WorkerHandler:             workerHandler,
		ContractorHandler:         contractorHandler,
		ReportHandler:             reportHandler,
		PermitHandler:             permitHandler,
		InspectionHandler:         inspectionHandler,
		IncidentHandler:           incidentHandler,
		EPPHandler:                eppHandler,
		EvacuationHandler:         evacuationHandler,
		TrainingHandler:           trainingHandler,
		OccupationalHealthHandler: occupationalHealthHandler,
		RealtimeHandler:           realtimeHandler,
		JWTSecret:                 cfg.JWTSecret,
		CORSOrigins:               cfg.CORSAllowedOrigins,
		LoginRateLimit:            loginLimit,
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
