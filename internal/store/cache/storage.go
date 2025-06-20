package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	RedisRateLimit interface {
		GetCount(ctx context.Context, key string) (int, error)
		Increment(ctx context.Context, key string) (int, error)
		TTL(ctx context.Context, key string) (time.Duration, error)
	}
}

func NewRedisStore(rdb *redis.Client) Storage {
	return Storage{
		RedisRateLimit: &RateLimitRedisStore{rdb: rdb},
	}
}
