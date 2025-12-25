//go:build gcloud

package main

import (
	"context"
	"log/slog"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/pubsub"
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
