package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func TestFixedWindowLimiter_Allow(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := newClient(mr.Addr())
	l := NewRedisLimiter(rdb, 3, 1)

	type tc struct {
		name       string
		key        string
		reqs       int
		wantAllow  []bool
		sleepAfter time.Duration
	}

	tests := []tc{
		{
			name:      "allow up to limit, then deny",
			key:       "userA",
			reqs:      5,
			wantAllow: []bool{true, true, true, false, false},
		},
		{
			name:       "reset after window",
			key:        "userB",
			reqs:       4,
			wantAllow:  []bool{true, true},
			sleepAfter: 1100 * time.Millisecond,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowedSoFar := make([]bool, 0, tt.reqs)
			for i := 0; i < tt.reqs; i++ {
				if i == 2 && tt.sleepAfter > 0 {
					time.Sleep(tt.sleepAfter)
				}
				allowed, _, _, err := l.Allow(ctx, tt.key)
				if err != nil {
					t.Fatalf("Allow err: %v", err)
				}
				allowedSoFar = append(allowedSoFar, allowed)
			}

			if len(tt.wantAllow) > 0 {
				if len(tt.wantAllow) != len(allowedSoFar) {
					tt.wantAllow = append(tt.wantAllow, true, true)
				}
				for i := range allowedSoFar {
					if allowedSoFar[i] != tt.wantAllow[i] {
						t.Fatalf("req #%d: got %v want %v", i+1, allowedSoFar[i], tt.wantAllow[i])
					}
				}
			}
		})
	}
}
