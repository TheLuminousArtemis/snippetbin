package ratelimiter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/theluminousartemis/snippetbin/internal/store/cache"
)

type RedisFixedWindowRateLimiter struct {
	store  cache.Storage
	limit  int
	window time.Duration
}

func NewRedisFixedWindowRateLimiter(store cache.Storage, limit int, window time.Duration) *RedisFixedWindowRateLimiter {
	return &RedisFixedWindowRateLimiter{
		store:  store,
		limit:  limit,
		window: window,
	}
}

func (r *RedisFixedWindowRateLimiter) Allow(ctx context.Context, ip string) (bool, time.Duration, error) {
	slog.Info(ip)
	key := fmt.Sprintf("ratelimit:%s", ip)

	count, err := r.store.RedisRateLimit.Increment(ctx, key)
	if err != nil {
		return false, 0, err
	}

	if count > r.limit {
		ttl, err := r.store.RedisRateLimit.TTL(ctx, key)
		if err != nil {
			return false, 0, err
		}
		return false, ttl, nil
	}

	return true, 0, nil
}
