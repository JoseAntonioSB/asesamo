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
	JWTSecret          string
	CORSOrigins        []string
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

	// Public Health endpoints (at root and /api/v1)
	r.Get("/health", cfg.HealthHandler.Check)
	r.Get("/api/v1/health", cfg.HealthHandler.Check)

	// API v1 group
	r.Route("/api/v1", func(r chi.Router) {
		// Rate limiter on login: 10 requests per minute per IP
		loginLimiter := httprate.LimitBy(
			10,
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
		})
	})

	return r
}
