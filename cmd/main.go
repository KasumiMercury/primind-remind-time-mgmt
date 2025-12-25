package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/handler"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize default logger for early logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)

		return 1
	}

	// Setup logger with configured level
	setupLogger(cfg.Log)

	// Validate pubsub configuration
	if err := cfg.PubSub.Validate(); err != nil {
		slog.Error("pubsub configuration error", "error", err)

		return 1
	}

	// Create cancellable context for cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)

		return 1
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get underlying sql.DB", "error", err)

		return 1
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Initialize publisher
	publisher, err := initPublisher(ctx, cfg)
	if err != nil {
		slog.Error("failed to create publisher", "error", err)

		return 1
	}

	if publisher != nil {
		defer func() {
			if err := publisher.Close(); err != nil {
				slog.Warn("failed to close publisher", "error", err)
			}
		}()
	}

	// Create repository, use case, and handler
	remindRepo := repository.NewRemindRepository(db)
	remindUseCase := app.NewRemindUseCase(remindRepo, publisher)
	remindHandler := handler.NewRemindHandler(remindUseCase)

	// Setup router
	router := setupRouter(remindHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)

	go func() {
		slog.Info("starting server", "address", cfg.Server.Address())

		serverErr <- srv.ListenAndServe()
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server", "error", err)

			return 1
		}

		slog.Info("server exited properly")

		return 0

	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}

		slog.Error("server exited with error", "error", err)

		return 1
	}
}

func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func setupRouter(remindHandler *handler.RemindHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := router.Group("/api/v1")
	remindHandler.RegisterRoutes(v1)

	return router
}

func setupLogger(cfg config.LogConfig) {
	var level slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
