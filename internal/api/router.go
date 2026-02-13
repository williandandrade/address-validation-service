package api

import (
	"net/http"

	"github.com/williandandrade/address-validation-service/internal/api/handler"
)

// Router holds all HTTP handlers and provides route registration.
type Router struct {
	mux             *http.ServeMux
	healthHandler   *handler.HealthHandler
	validateHandler *handler.ValidateHandler
}

// NewRouter creates a new Router with the given handlers.
func NewRouter(
	healthHandler *handler.HealthHandler,
	validateHandler *handler.ValidateHandler,
) *Router {
	return &Router{
		mux:             http.NewServeMux(),
		healthHandler:   healthHandler,
		validateHandler: validateHandler,
	}
}

// Setup registers all routes and returns the handler.
func (r *Router) Setup() http.Handler {
	// Health endpoints (no authentication required)
	r.mux.HandleFunc("GET /api/v1/health", r.healthHandler.Health)
	r.mux.HandleFunc("GET /api/v1/health/ready", r.healthHandler.Ready)

	// Validation endpoints
	if r.validateHandler != nil {
		r.mux.HandleFunc("POST /api/v1/validate-address", r.validateHandler.Validate)
	}

	return r.mux
}

// Handler returns the configured HTTP handler.
func (r *Router) Handler() http.Handler {
	return r.Setup()
}
