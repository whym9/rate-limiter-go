package redis

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	rdb        *redis.Client
	capacity   int
	refillRate float64
	ttl        time.Duration
	prefix     string
}

func NewTokenBucketLimiter(rdb *redis.Client, capacity int, windowSec int, keyPrefix string) *TokenBucketLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if windowSec <= 0 {
		windowSec = 60
	}
	return &TokenBucketLimiter{
		rdb:        rdb,
		capacity:   capacity,
		refillRate: float64(capacity) / float64(windowSec),
		ttl:        time.Duration(windowSec*2) * time.Second,
		prefix:     keyPrefix,
	}
}

func (l *TokenBucketLimiter) keyFor(client string) string {
	if l.prefix == "" {
		return fmt.Sprintf("rl:tb:%s", client)
	}
	return fmt.Sprintf("%s:tb:%s", l.prefix, client)
}

var tokenBucketLua = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now  = tonumber(ARGV[3])
local ttl  = tonumber(ARGV[4])

local vals = redis.call('HMGET', key, 'tokens', 'ts')
local tokens = tonumber(vals[1])
local ts = tonumber(vals[2])

if tokens == nil then
  tokens = capacity
  ts = now
else
  local delta = now - ts
  if delta < 0 then delta = 0 end
  tokens = math.min(capacity, tokens + (delta * rate))
  if delta > 0 then ts = now end
end

local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call('HMSET', key, 'tokens', tokens, 'ts', ts)
redis.call('EXPIRE', key, ttl)

return {allowed, tokens, ts}
`)

func (l *TokenBucketLimiter) Allow(ctx context.Context, clientKey string) (bool, int, time.Time, error) {
	now := time.Now().UTC()
	nowSec := now.Unix()

	res, err := tokenBucketLua.Run(ctx, l.rdb,
		[]string{l.keyFor(clientKey)},
		l.capacity,
		l.refillRate,
		nowSec,
		int64(l.ttl.Seconds()),
	).Result()
	if err != nil {
		return false, 0, now, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 3 {
		return false, 0, now, fmt.Errorf("unexpected lua result: %#v", res)
	}

	allowedInt := toInt64(arr[0])
	tokens := toFloat(arr[1])
	allowed := allowedInt == 1
	remaining := int(math.Floor(tokens))

	var need float64
	if !allowed {
		need = 1.0 - tokens
		if need < 0 {
			need = 0
		}
	} else {
		need = float64(l.capacity) - tokens
		if need < 0 {
			need = 0
		}
	}
	secs := int64(math.Ceil(need / l.refillRate))
	resetAt := time.Unix(nowSec+secs, 0).UTC()

	return allowed, remaining, resetAt, nil
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	case int64:
		return float64(t)
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case float64:
		return int64(t)
	case string:
		i, _ := strconv.ParseInt(t, 10, 64)
		return i
	default:
		return 0
	}
}
