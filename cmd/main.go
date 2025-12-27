package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/handler"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/logging"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	obs, err := initObservability(ctx)
	if err != nil {
		slog.Error("failed to initialize observability", slog.String("error", err.Error()))

		return err
	}

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := obs.Shutdown(shutdownCtx); err != nil {
			slog.Warn("failed to shutdown observability", slog.String("error", err.Error()))
		}
	}()

	slog.SetDefault(obs.Logger())

	cfg, err := config.Load()
	if err != nil {
		slog.ErrorContext(ctx, "failed to load configuration",
			slog.String("event", "config.load.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	// Validate pubsub configuration
	if err := cfg.PubSub.Validate(); err != nil {
		slog.ErrorContext(ctx, "pubsub configuration error",
			slog.String("event", "config.validate.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	// Create cancellable context for cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize database",
			slog.String("event", "db.init.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get underlying sql.DB",
			slog.String("event", "db.handle.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Warn("failed to close database connection", slog.String("error", err.Error()))
		}
	}()

	publisher, err := initPublisher(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create publisher",
			slog.String("event", "pubsub.init.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	var closePublisherOnce sync.Once

	closePublisher := func() {
		closePublisherOnce.Do(func() {
			if publisher != nil {
				if err := publisher.Close(); err != nil {
					slog.Warn("failed to close publisher", slog.String("error", err.Error()))
				}
			}
		})
	}
	defer closePublisher()

	// Create repository, use case, and handler
	remindRepo := repository.NewRemindRepository(db)
	remindUseCase := app.NewRemindUseCase(remindRepo, publisher)
	remindHandler := handler.NewRemindHandler(remindUseCase)

	// Setup router
	router := setupRouter(remindHandler)

	server := &http.Server{
		Addr:              cfg.Server.Address(),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.InfoContext(ctx, "starting server",
		slog.String("event", "server.start"),
		slog.String("address", cfg.Server.Address()),
		slog.String("version", Version),
	)

	go func() {
		<-ctx.Done()

		slog.InfoContext(ctx, "shutdown signal received",
			slog.String("event", "server.shutdown.start"),
		)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server", slog.String("error", err.Error()))
		}

		closePublisher()
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.ErrorContext(ctx, "server exited with error",
			slog.String("event", "server.exit.fail"),
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.InfoContext(ctx, "server stopped",
		slog.String("event", "server.stop"),
	)

	return nil
}

func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logging.NewGormLogger(200 * time.Millisecond),
	})
	if err != nil {
		return nil, err
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to register GORM tracing plugin: %w", err)
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
