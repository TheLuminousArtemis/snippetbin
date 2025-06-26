package cache

import (
	"context"
	"time"
)

func NewMockStorage() Storage {
	return Storage{
		RedisRateLimit: &MockRedisRateLimit{},
	}
}

type MockRedisRateLimit struct {
	count int
}

func (m *MockRedisRateLimit) Increment(ctx context.Context, key string) (int, error) {
	m.count++
	return m.count, nil
}

func (m *MockRedisRateLimit) TTL(ctx context.Context, key string) (time.Duration, error) {
	return time.Second, nil
}
