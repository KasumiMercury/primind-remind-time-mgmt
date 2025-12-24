//go:build !gcloud

package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	nc "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	throttlev1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/throttle/v1"
	pjson "github.com/KasumiMercury/primind-remind-time-mgmt/internal/proto"
)

type NATSPublisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

type NATSPublisherConfig struct {
	URL string
}

func NewNATSPublisher(cfg NATSPublisherConfig) (*NATSPublisher, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL:         cfg.URL,
			NatsOptions: []nc.Option{nc.Timeout(10 * time.Second)},
			JetStream: nats.JetStreamConfig{
				Disabled:       false,
				AutoProvision:  true,
				ConnectOptions: nil,
				PublishOptions: nil,
				TrackMsgId:     false,
				AckAsync:       false,
				DurablePrefix:  "",
			},
			Marshaler: &nats.NATSMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS publisher: %w", err)
	}

	return &NATSPublisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

func NewNATSPublisherWithStream(ctx context.Context, cfg NATSPublisherConfig) (*NATSPublisher, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	conn, err := nc.Connect(cfg.URL, nc.Timeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer conn.Close()

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	streamName := "REMIND_EVENTS"

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        streamName,
		Description: "Stream for remind events",
		Subjects:    []string{TopicRemindCancelled},
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      24 * time.Hour,
		MaxBytes:    100 * 1024 * 1024, // 100MB
		Storage:     jetstream.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	slog.Info("NATS JetStream stream configured",
		slog.String("stream", streamName),
		slog.String("subject", TopicRemindCancelled),
	)

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL:         cfg.URL,
			NatsOptions: []nc.Option{nc.Timeout(10 * time.Second)},
			JetStream: nats.JetStreamConfig{
				Disabled:      false,
				AutoProvision: false,
			},
			Marshaler: &nats.NATSMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS publisher: %w", err)
	}

	return &NATSPublisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

func (p *NATSPublisher) PublishRemindCancelled(ctx context.Context, req *throttlev1.CancelRemindRequest) error {
	payload, err := pjson.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("event_type", "remind.cancelled")
	msg.Metadata.Set("task_id", req.GetTaskId())
	msg.Metadata.Set("user_id", req.GetUserId())

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

func (p *NATSPublisher) Close() error {
	return p.publisher.Close()
}
