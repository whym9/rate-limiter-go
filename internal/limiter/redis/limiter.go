package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	rdb       *redis.Client
	limit     int
	windowDur time.Duration
}

func NewRedisLimiter(rdb *redis.Client, limit int, windowSec int) *RedisLimiter {
	if windowSec <= 0 {
		windowSec = 1
	}
	return &RedisLimiter{
		rdb:       rdb,
		limit:     limit,
		windowDur: time.Duration(windowSec) * time.Second,
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, int, time.Time, error) {
	now := time.Now().UTC()
	windowStart := now.Truncate(l.windowDur)
	windowEnd := windowStart.Add(l.windowDur)

	redisKey := fmt.Sprintf("rl:%s:%d", key, windowStart.Unix())

	ttl := time.Until(windowEnd)
	if ttl <= 0 {
		ttl = l.windowDur
	}

	var count int64

	count, err := l.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, 0, windowEnd, err
	}
	if count == 1 {
		_ = l.rdb.Expire(ctx, redisKey, ttl).Err()
	}

	allowed := count <= int64(l.limit)
	remaining := l.limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return allowed, remaining, windowEnd, nil
}
