package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/api/v1/router"
	repository "github.com/ciaranmcdonnell/go-api-server/internal/core/common/repository"
	service "github.com/ciaranmcdonnell/go-api-server/internal/core/common/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/database"
	"github.com/ciaranmcdonnell/go-api-server/internal/database/migrations"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func Execute() {
	config, err := utils.GetConfig()
	if err != nil {
		slog.Error("Could not load config", slog.Any("error", err))
		os.Exit(1)
	}

	runMigrations := migrations.ShouldRunMigrations()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := database.InitializeConnectionPool(ctx, config, runMigrations); err != nil {
		slog.Error("Could not connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer database.CloseDB()

	var queriesManager repository.QueriesInterface = repository.NewQueries(database.DBPool)
	var servicesManager service.ServicesInterface = service.NewServices(config, queriesManager)

	r := router.Setup(config, servicesManager, queriesManager)

	srv := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "address", config.ServerAddress, "environment", config.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	slog.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	}

	cancel()
	slog.Info("Server gracefully stopped")
}
