package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/handler"
	"github.com/dnakolan/worker-pool-service/internal/pool"
	"github.com/dnakolan/worker-pool-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	healthHandler := handler.NewHealthHandler()
	router.Get("/health", healthHandler.GetHealthHandler)

	pool := pool.NewWorkerPool(context.Background(), 10, 10)
	pool.Start()
	defer pool.Stop()

	jobService := service.NewJobsService(pool)
	jobsHandler := handler.NewJobsHandler(jobService)

	router.Post("/jobs", jobsHandler.CreateJobsHandler)
	router.Get("/jobs", jobsHandler.ListJobsHandler)
	router.Get("/jobs/{uid}", jobsHandler.GetJobsHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("Received terminate, graceful shutdown", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server Shutdown Failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Server exited properly")

	os.Exit(0)
}
