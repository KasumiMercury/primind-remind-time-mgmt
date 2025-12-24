package pubsub

import (
	"context"
	"io"

	throttlev1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/throttle/v1"
)

//go:generate mockgen -source=publisher.go -destination=publisher_mock.go -package=pubsub

type Publisher interface {
	PublishRemindCancelled(ctx context.Context, req *throttlev1.CancelRemindRequest) error
	io.Closer
}
