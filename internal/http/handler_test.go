package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	logi "rate-limiter-go/internal/log"
)

// --- a tiny fake limiter to drive responses ---
type fakeLimiter struct {
	allowNext bool
}

func (f *fakeLimiter) Allow(_ context.Context, _ string) (bool, int, time.Time, error) {
	if f.allowNext {
		return true, 99, time.Now().Add(time.Second), nil
	}
	return false, 0, time.Now().Add(time.Second), nil
}

func TestHealthz(t *testing.T) {
	logger := logi.NewLogger()
	srv := New(logger, ":9090")
	RegisterEndpoints(logger, nil, srv.Router())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("want 200 got %d", res.StatusCode)
	}
}

func TestRateLimit(t *testing.T) {
	type tc struct {
		name       string
		allow      bool
		wantStatus int
		authHeader string
	}
	tests := []tc{
		{"allow path", true, http.StatusOK, "userA"},
		{"deny path", false, http.StatusTooManyRequests, "userA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := &fakeLimiter{allowNext: tt.allow}
			logger := logi.NewLogger()
			srv := New(logger, ":9090")
			RegisterEndpoints(logger, fl, srv.Router())
			ts := httptest.NewServer(srv.Router())
			defer ts.Close()

			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/rate-limit", strings.NewReader(`{}`))
			req.Header.Set("Authorization", tt.authHeader)
			req.Header.Set("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("POST /v1/rate-limit: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status: got %d want %d", res.StatusCode, tt.wantStatus)
			}

			var body struct {
				Allowed   bool      `json:"allowed"`
				Remaining int       `json:"remaining"`
				ResetAt   time.Time `json:"reset_at"`
			}
			_ = json.NewDecoder(res.Body).Decode(&body)
			if body.Allowed != tt.allow {
				t.Fatalf("Allowed: got %v want %v", body.Allowed, tt.allow)
			}
		})
	}
}
