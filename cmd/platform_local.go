//go:build !gcloud

package main

import (
	"context"
	"log/slog"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/pubsub"
)

func initPublisher(ctx context.Context, cfg *config.Config) (pubsub.Publisher, error) {
	if cfg.PubSub.NatsURL == "" {
		slog.Warn("NATS_URL not set, event publishing disabled")
		return nil, nil
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
