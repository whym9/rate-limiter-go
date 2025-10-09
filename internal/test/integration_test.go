//go:build integration
// +build integration

package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	tc "github.com/testcontainers/testcontainers-go"
	credis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	myhttp "rate-limiter-go/internal/http"
	rl "rate-limiter-go/internal/limiter/redis"
	logi "rate-limiter-go/internal/log"
)

func TestEndToEnd_WithRealRedis(t *testing.T) {
	ctx := context.Background()

	// spin up Redis container
	rC, err := credis.Run(ctx, "redis:7-alpine",
		tc.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	if err != nil {
		t.Fatalf("redis container: %v", err)
	}
	t.Cleanup(func() { _ = rC.Terminate(ctx) })

	host, _ := rC.Host(ctx)
	port, _ := rC.MappedPort(ctx, "6379/tcp")
	addr := host + ":" + port.Port()

	// real Redis client + real limiter
	rdb := redis2(addr) // small helper that returns *redis.Client
	l := rl.NewRedisLimiter(rdb, 3, 1)

	// real HTTP server (in-memory) using your router
	logger := logi.NewLogger()
	srv := myhttp.New(logger, ":9090" /* plus any deps */)
	myhttp.RegisterEndpoints(logger, l, srv.Router())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	// hit healthz
	res, err := http.Get(ts.URL + "/healthz")
	if err != nil || res.StatusCode != 200 {
		t.Fatalf("healthz: %v status=%d", err, res.StatusCode)
	}

	// hit rate-limit 4 times, expect 3 OK then 429
	for i := 1; i <= 4; i++ {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/rate-limit", nil)
		req.Header.Set("Authorization", "user1")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("req %d: %v", i, err)
		}
		if i <= 3 && resp.StatusCode != 200 {
			t.Fatalf("req %d want 200 got %d", i, resp.StatusCode)
		}
		if i == 4 && resp.StatusCode != 429 {
			t.Fatalf("req 4 want 429 got %d", resp.StatusCode)
		}
		var body struct {
			Allowed bool `json:"allowed"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
	}
}

// helper to create redis client
func redis2(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}
