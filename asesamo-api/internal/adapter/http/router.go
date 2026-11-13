package http

import (
	"net/http"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	customMiddleware "asesamo-api/internal/adapter/http/middleware"
	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// RouterConfig contains dependencies for configuring the Chi router.
type RouterConfig struct {
	HealthHandler      *handler.HealthHandler
	AuthHandler        *handler.AuthHandler
	AccessPointHandler *handler.AccessPointHandler
	AccessHandler      *handler.AccessHandler
	DocumentHandler    *handler.DocumentHandler
	UserHandler        *handler.UserHandler
	WorkerHandler      *handler.WorkerHandler
	ContractorHandler  *handler.ContractorHandler
	ReportHandler      *handler.ReportHandler
	PermitHandler             *handler.PermitHandler
	InspectionHandler         *handler.InspectionHandler
	IncidentHandler           *handler.IncidentHandler
	EPPHandler                *handler.EPPHandler
	EvacuationHandler         *handler.EvacuationHandler
	TrainingHandler           *handler.TrainingHandler
	OccupationalHealthHandler *handler.OccupationalHealthHandler
	RealtimeHandler           *handler.RealtimeHandler
	JWTSecret                 string
	CORSOrigins        []string
	LoginRateLimit     int
}

// NewRouter constructs and configures the HTTP request multiplexer.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(customMiddleware.CORS(cfg.CORSOrigins))

	// Public Health endpoints (at root, /healthz, and /api/v1)
	r.Get("/health", cfg.HealthHandler.Check)
	r.Get("/healthz", cfg.HealthHandler.Check)
	r.Get("/api/v1/health", cfg.HealthHandler.Check)

	// API v1 group
	r.Route("/api/v1", func(r chi.Router) {
		limit := cfg.LoginRateLimit
		if limit <= 0 {
			limit = 10
		}
		// Rate limiter on login: requests per minute per IP
		loginLimiter := httprate.LimitBy(
			limit,
			1*time.Minute,
			httprate.KeyByIP,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				dto.WriteError(w, domainErrors.ErrTooManyRequests)
			}),
		)

		// Public Auth
		r.With(loginLimiter).Post("/auth/login", cfg.AuthHandler.Login)

		// Public Avatar route (for <img> tags in badges and cards)
		if cfg.UserHandler != nil {
			r.Get("/users/{id}/avatar", cfg.UserHandler.GetAvatar)
		}

		// Protected routes (JWT authenticated)
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Auth(cfg.JWTSecret))

			// Access point catalog
			r.Get("/access-points", cfg.AccessPointHandler.List)

			// Gate Access Verification, Check-Out, and Audit Logs
			if cfg.AccessHandler != nil {
				r.Post("/access/check-in", cfg.AccessHandler.CheckIn)
				r.Post("/access/check-out", cfg.AccessHandler.CheckOut)
				r.Get("/access-logs", cfg.AccessHandler.ListLogs)
			}

			// Document Management & EHS Review
			if cfg.DocumentHandler != nil {
				r.Post("/documents/upload", cfg.DocumentHandler.Upload)
				r.Get("/documents/my-documents", cfg.DocumentHandler.ListMyDocuments)
				r.Get("/documents/pending", cfg.DocumentHandler.ListPending)
				r.Post("/documents/{id}/review", cfg.DocumentHandler.Review)
				r.Get("/documents/{id}/download", cfg.DocumentHandler.Download)
				r.Get("/documents/types", cfg.DocumentHandler.ListTypes)
			}

			// User Profile & Preferences
			if cfg.UserHandler != nil {
				r.Get("/users/me", cfg.UserHandler.GetProfile)
				r.Patch("/users/me", cfg.UserHandler.UpdateProfile)
				r.Patch("/users/me/theme", cfg.UserHandler.UpdateTheme)
				r.Post("/users/me/avatar", cfg.UserHandler.UploadMyAvatar)
				r.Post("/users/{id}/avatar", cfg.UserHandler.UploadUserAvatar)
			}

			// Worker Credential & Mobile Badge
			if cfg.WorkerHandler != nil {
				r.Get("/workers/me/credential", cfg.WorkerHandler.GetMyCredential)
				r.Get("/workers/{id}/credential", cfg.WorkerHandler.GetWorkerCredential)
			}

			// Contractor & Crew Management (Phase 2 Module 1)
			if cfg.ContractorHandler != nil {
				r.Route("/contractors", func(cr chi.Router) {
					// Contractor Companies
					cr.With(customMiddleware.RequireRole("plant_admin", "ehs_manager", "tenant_admin")).Post("/invite", cfg.ContractorHandler.InviteContractor)
					cr.Get("/", cfg.ContractorHandler.ListPlantContractors)
					cr.With(customMiddleware.RequireRole("plant_admin", "ehs_manager", "tenant_admin")).Patch("/{id}/status", cfg.ContractorHandler.UpdateContractorStatus)

					// Contractor Self Company Info
					cr.With(customMiddleware.RequireRole("contractor_admin")).Get("/my-company", cfg.ContractorHandler.GetMyCompany)

					// Crew Management
					cr.Get("/workers", cfg.ContractorHandler.ListWorkers)
					cr.With(customMiddleware.RequireRole("contractor_admin", "plant_admin", "ehs_manager", "tenant_admin")).Post("/workers", cfg.ContractorHandler.AddWorker)
					cr.With(customMiddleware.RequireRole("contractor_admin", "plant_admin", "ehs_manager", "tenant_admin")).Patch("/workers/{id}/status", cfg.ContractorHandler.ToggleWorkerStatus)
				})
			}

			// Metrics & STPS Audit Reports (Phase 2 Module 2)
			if cfg.ReportHandler != nil {
				r.Route("/reports", func(rr chi.Router) {
					rr.Use(customMiddleware.RequireRole("plant_admin", "ehs_manager", "tenant_admin", "guard", "contractor_admin"))
					rr.Get("/metrics", cfg.ReportHandler.GetMetrics)
					rr.Get("/export/access-logs", cfg.ReportHandler.ExportAccessLogs)
					rr.Get("/export/compliance", cfg.ReportHandler.ExportCompliance)
				})
			}

			// High-Risk Work Permits (PTAR Digital)
			if cfg.PermitHandler != nil {
				r.Route("/permits", func(pr chi.Router) {
					pr.Get("/", cfg.PermitHandler.ListPermits)
					pr.With(customMiddleware.RequireRole("contractor_admin", "plant_admin", "ehs_manager", "tenant_admin", "super_admin")).Post("/", cfg.PermitHandler.CreatePermit)
					pr.Get("/{id}", cfg.PermitHandler.GetPermit)
					pr.With(customMiddleware.RequireRole("contractor_admin", "plant_admin", "ehs_manager", "tenant_admin", "super_admin")).Post("/{id}/workers", cfg.PermitHandler.AssignWorkers)
					pr.With(customMiddleware.RequireRole("contractor_admin", "plant_admin", "ehs_manager", "tenant_admin", "super_admin")).Patch("/{id}/checklist", cfg.PermitHandler.UpdateChecklist)
					pr.With(customMiddleware.RequireRole("ehs_manager", "plant_admin", "tenant_admin", "super_admin")).Post("/{id}/approve", cfg.PermitHandler.ApprovePermit)
					pr.Post("/{id}/status", cfg.PermitHandler.TransitionStatus)
				})
			}

			// Inspection & EHS Audit endpoints
			if cfg.InspectionHandler != nil {
				r.Route("/inspections", func(ir chi.Router) {
					ir.Get("/templates", cfg.InspectionHandler.ListTemplates)
					ir.Get("/", cfg.InspectionHandler.ListInspections)
					ir.Post("/", cfg.InspectionHandler.CreateInspection)
					ir.Post("/{id}/complete", cfg.InspectionHandler.CompleteInspection)
				})
			}

			// Incident & Near-Miss endpoints (NOM-021-STPS)
			if cfg.IncidentHandler != nil {
				r.Route("/incidents", func(inr chi.Router) {
					inr.Get("/", cfg.IncidentHandler.ListIncidents)
					inr.Post("/", cfg.IncidentHandler.ReportIncident)
					inr.Get("/metrics", cfg.IncidentHandler.GetMetrics)
				})
			}

			// EPP Management & Matrix endpoints (NOM-017-STPS)
			if cfg.EPPHandler != nil {
				r.Route("/epp", func(er chi.Router) {
					er.Get("/catalog", cfg.EPPHandler.ListCatalog)
					er.Post("/catalog", cfg.EPPHandler.CreateItem)
					er.Get("/deliveries", cfg.EPPHandler.ListDeliveries)
					er.Post("/deliveries", cfg.EPPHandler.DeliverEPP)
				})
			}

			// Emergency Evacuation & Roll Call endpoints (NOM-002-STPS)
			if cfg.EvacuationHandler != nil {
				r.Route("/evacuations", func(evr chi.Router) {
					evr.Get("/muster-points", cfg.EvacuationHandler.ListMusterPoints)
					evr.Get("/active", cfg.EvacuationHandler.GetActiveEvacuation)
					evr.Post("/start", cfg.EvacuationHandler.StartEvacuation)
					evr.Post("/{id}/check-in", cfg.EvacuationHandler.CheckInWorker)
					evr.Post("/{id}/end", cfg.EvacuationHandler.EndEvacuation)
				})
			}

			// Training & STPS DC-3 / DC-4 endpoints
			if cfg.TrainingHandler != nil {
				r.Route("/training", func(tr chi.Router) {
					tr.Get("/courses", cfg.TrainingHandler.ListCourses)
					tr.Post("/courses", cfg.TrainingHandler.CreateCourse)
					tr.Get("/dc3", cfg.TrainingHandler.ListCertificates)
					tr.Post("/dc3", cfg.TrainingHandler.IssueDC3)
					tr.Get("/workers/{workerId}/dc3", cfg.TrainingHandler.ListWorkerCertificates)
				})
			}

			// Occupational Health & Medical Exams (NOM-030-STPS)
			if cfg.OccupationalHealthHandler != nil {
				r.Route("/health", func(hr chi.Router) {
					hr.Get("/workers/{workerId}/profile", cfg.OccupationalHealthHandler.GetWorkerMedicalProfile)
					hr.Put("/workers/{workerId}/profile", cfg.OccupationalHealthHandler.UpsertWorkerMedicalProfile)
					hr.Get("/exams", cfg.OccupationalHealthHandler.ListAllExams)
					hr.Post("/exams", cfg.OccupationalHealthHandler.RecordMedicalExam)
					hr.Get("/workers/{workerId}/exams", cfg.OccupationalHealthHandler.ListWorkerExams)
				})
			}
			// Real-Time SSE Event Stream
			if cfg.RealtimeHandler != nil {
				r.Route("/realtime", func(rr chi.Router) {
					rr.Get("/events", cfg.RealtimeHandler.StreamEvents)
				})
			}
		})
	})

	return r
}
