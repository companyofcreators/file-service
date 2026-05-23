package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/companyofcreators/file-service/internal/app"
	httphandler "github.com/companyofcreators/file-service/internal/interfaces/http"
	"github.com/companyofcreators/file-service/pkg/header_auth"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		panic("failed to initialize container: " + err.Error())
	}
	defer container.Shutdown()

	logger := container.Logger

	headerSigner := header_auth.NewHeaderSigner(container.Config.HeaderHMACKey)
	router := httphandler.NewRouter(container.Handler, headerSigner, logger, container.Config.AllowedOrigin)

	srv := &http.Server{
		Addr:         container.Config.HTTPAddress,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("file-service starting", slog.String("address", container.Config.HTTPAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server exited gracefully")
}
