package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/companyofcreators/file-service/pkg/header_auth"
)

func NewRouter(h *FileHandler, signer *header_auth.HeaderSigner, logger *slog.Logger, allowedOrigin string) *chi.Mux {
	r := chi.NewRouter()

	// Built-in chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Custom middleware
	r.Use(loggingMiddleware(logger))
	r.Use(corsMiddleware(allowedOrigin))
	r.Use(signer.VerifyMiddleware)

	// File endpoints
	r.Post("/internal/files/upload", h.Upload)
	r.Get("/internal/files/{id}", h.GetFile)
	r.Get("/internal/files/{id}/download", h.Download)
	r.Delete("/internal/files/{id}", h.Delete)
	r.Get("/internal/files", h.ListFiles)
	r.Get("/internal/health", h.Health)

	return r
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := allowedOrigin
			if origin == "" {
				origin = "http://localhost:8080"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-User-Role")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
