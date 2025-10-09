package limiter

import (
	"context"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, int, time.Time, error)
}
