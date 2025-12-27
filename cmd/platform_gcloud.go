//go:build gcloud

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
	publisher, err := pubsub.NewGCloudPublisher(ctx, pubsub.GCloudPublisherConfig{
		ProjectID: cfg.PubSub.GCloudProjectID,
	})
	if err != nil {
		return nil, err
	}

	slog.Info("Google Cloud Pub/Sub publisher initialized",
		"project_id", cfg.PubSub.GCloudProjectID,
	)

	return publisher, nil
}

func initObservability(ctx context.Context) (*observability.Resources, error) {
	serviceName := os.Getenv("K_SERVICE")
	if serviceName == "" {
		serviceName = "time-mgmt"
	}

	env := logging.EnvProd
	if e := os.Getenv("ENV"); e != "" {
		env = logging.Environment(e)
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCLOUD_PROJECT_ID")
	}

	obs, err := observability.Init(ctx, observability.Config{
		ServiceInfo: logging.ServiceInfo{
			Name:     serviceName,
			Version:  Version,
			Revision: os.Getenv("K_REVISION"),
		},
		Environment:   env,
		GCPProjectID:  projectID,
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
