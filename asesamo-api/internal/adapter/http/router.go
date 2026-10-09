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
	JWTSecret          string
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
				r.Patch("/users/me/theme", cfg.UserHandler.UpdateTheme)
			}

			// Worker Credential & Mobile Badge
			if cfg.WorkerHandler != nil {
				r.Get("/workers/me/credential", cfg.WorkerHandler.GetMyCredential)
				r.Get("/workers/{id}/credential", cfg.WorkerHandler.GetWorkerCredential)
			}
		})
	})

	return r
}
