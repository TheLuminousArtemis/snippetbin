package ratelimiter

import (
	"context"
	"time"
)

// type Config struct {
// 	RequestsPerTimeFrame int
// 	Timeframe            time.Duration
// 	Enabled              bool
// }

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, time.Duration, error)
}
