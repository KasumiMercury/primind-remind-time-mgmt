//go:build !gcloud

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/handler"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/pubsub"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/logging"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/metrics"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/middleware"
)

func initPublisher(ctx context.Context, cfg *config.Config) (pubsub.Publisher, error) {
	if cfg.PubSub.NatsURL == "" {
		slog.Warn("NATS_URL not set, event publishing disabled")

		return nil, nil //nolint:nilnil
	}

	publisher, err := pubsub.NewNATSPublisherWithStream(ctx, pubsub.NATSPublisherConfig{
		URL: cfg.PubSub.NatsURL,
	})
	if err != nil {
		return nil, err
	}

	slog.Info("NATS publisher initialized", "url", cfg.PubSub.NatsURL)

	return publisher, nil
}

func initObservability(ctx context.Context) (*observability.Resources, error) {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "time-mgmt"
	}

	env := logging.EnvDev
	if e := os.Getenv("ENV"); e != "" {
		env = logging.Environment(e)
	}

	obs, err := observability.Init(ctx, observability.Config{
		ServiceInfo: logging.ServiceInfo{
			Name:     serviceName,
			Version:  Version,
			Revision: "",
		},
		Environment:   env,
		GCPProjectID:  "",
		SamplingRate:  1.0,
		DefaultModule: logging.Module("remind"),
	})
	if err != nil {
		return nil, err
	}

	return obs, nil
}

func setupRouter(remindHandler *handler.RemindHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	httpMetrics, err := metrics.NewHTTPMetrics()
	if err != nil {
		slog.Warn("failed to initialize HTTP metrics",
			slog.String("error", err.Error()),
		)
	}

	// Add observability middleware
	router.Use(
		middleware.Gin(middleware.GinConfig{
			SkipPaths:       []string{"/ping", "/health", "/healthz", "/metrics"},
			Module:          logging.Module("remind"),
			ModuleResolver:  nil,
			Worker:          false,
			JobNameResolver: nil,
			TracerName:      "github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/middleware",
			HTTPMetrics:     httpMetrics,
		}),
		middleware.PanicRecoveryGin(),
	)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := router.Group("/api/v1")
	remindHandler.RegisterRoutes(v1)

	return router
}
