package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitRedisStore struct {
	rdb *redis.Client
}

var window time.Duration = time.Minute * 2

// func (r *RateLimitRedisStore) GetCount(ctx context.Context, key string) (int, error) {
// 	count, err := r.rdb.Get(ctx, key).Int()
// 	if err == redis.Nil {
// 		return 0, nil
// 	}
// 	return count, err

// }
func (r *RateLimitRedisStore) Increment(ctx context.Context, key string) (int, error) {
	count, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := r.rdb.Expire(ctx, key, window).Err(); err != nil {
			return 0, err
		}
	}
	return int(count), nil

}
func (r *RateLimitRedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.rdb.TTL(ctx, key).Result()
}
