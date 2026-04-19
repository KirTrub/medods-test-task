package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	infrastructurepostgres "github.com/KirTrub/medods-test-task/internal/infrastructure/postgres"
	postgresrepo "github.com/KirTrub/medods-test-task/internal/repository/postgres"
	"github.com/KirTrub/medods-test-task/internal/scheduler"
	transporthttp "github.com/KirTrub/medods-test-task/internal/transport/http"
	swaggerdocs "github.com/KirTrub/medods-test-task/internal/transport/http/docs"
	httphandlers "github.com/KirTrub/medods-test-task/internal/transport/http/handlers"
	scheduleusecase "github.com/KirTrub/medods-test-task/internal/usecase/schedule"
	taskusecase "github.com/KirTrub/medods-test-task/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	scheduleRepo := postgresrepo.NewScheduleRepository(pool)

	taskSvc := taskusecase.NewService(taskRepo)
	scheduleSvc := scheduleusecase.NewService(scheduleRepo)

	taskHandler := httphandlers.NewTaskHandler(taskSvc)
	scheduleHandler := httphandlers.NewScheduleHandler(scheduleSvc)
	docsHandler := swaggerdocs.NewHandler()

	router := transporthttp.NewRouter(taskHandler, scheduleHandler, docsHandler)

	sched := scheduler.New(scheduleRepo, taskRepo, logger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sched.Run(ctx)
	}()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}

	wg.Wait()
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@postgres:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
