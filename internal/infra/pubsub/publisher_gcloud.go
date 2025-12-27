//go:build gcloud

package pubsub

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/v2/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"

	throttlev1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/throttle/v1"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/logging"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/observability/tracing"
	pjson "github.com/KasumiMercury/primind-remind-time-mgmt/internal/proto"
)

type GCloudPublisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

type GCloudPublisherConfig struct {
	ProjectID string
}

func NewGCloudPublisher(ctx context.Context, cfg GCloudPublisherConfig) (*GCloudPublisher, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	publisher, err := googlecloud.NewPublisher(
		googlecloud.PublisherConfig{
			ProjectID: cfg.ProjectID,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud publisher: %w", err)
	}

	return &GCloudPublisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

func (p *GCloudPublisher) PublishRemindCancelled(ctx context.Context, req *throttlev1.CancelRemindRequest) error {
	payload, err := pjson.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("message_type", "remind.cancel")
	msg.Metadata.Set("task_id", req.GetTaskId())
	msg.Metadata.Set("user_id", req.GetUserId())

	// Inject trace context into message metadata
	carrier := make(map[string]string)
	tracing.InjectToMap(ctx, carrier)
	for k, v := range carrier {
		msg.Metadata.Set(k, v)
	}
	reqID := logging.RequestIDFromContext(ctx)
	if reqID == "" {
		reqID = logging.ValidateAndExtractRequestID("")
	}
	msg.Metadata.Set("x-request-id", reqID)

	if err := p.publisher.Publish(TopicRemindCancelled, msg); err != nil {
		slog.Error("failed to publish remind cancelled event",
			slog.String("task_id", req.GetTaskId()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	slog.Debug("published remind cancelled event",
		slog.String("task_id", req.GetTaskId()),
		slog.String("message_id", msg.UUID),
	)
	return nil
}

func (p *GCloudPublisher) Close() error {
	return p.publisher.Close()
}
