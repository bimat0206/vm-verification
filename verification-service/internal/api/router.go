package api

import (
	//"net/http"

	"github.com/gorilla/mux"

	"verification-service/internal/api/handlers"
	"verification-service/internal/api/middleware"
)

// NewRouter creates a new HTTP router
func NewRouter(
	verificationHandler *handlers.VerificationHandler,
	healthHandler *handlers.HealthHandler,
) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.ContentTypeMiddleware)
	router.Use(middleware.CORSMiddleware)

	// API Routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	
	// Verification Routes
	apiRouter.HandleFunc("/verification", verificationHandler.InitiateVerification).Methods("POST")
	apiRouter.HandleFunc("/verification/{id}", verificationHandler.GetVerification).Methods("GET")
	apiRouter.HandleFunc("/verification", verificationHandler.ListVerifications).Methods("GET")
	
	// Health Routes
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")
	apiRouter.HandleFunc("/health/details", healthHandler.DetailedHealthCheck).Methods("GET")

	return router
}