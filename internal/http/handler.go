package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"rate-limiter-go/internal/limiter"
	logi "rate-limiter-go/internal/log"
	metrics "rate-limiter-go/internal/metrics"
)

type Handler struct {
	logger  *logi.Logger
	limiter limiter.Limiter
}

type Response struct {
	Key       string    `json:"-"`
	Allowed   bool      `json:"allowed"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

func newHandler(limiter limiter.Limiter, logger *logi.Logger) *Handler {
	return &Handler{
		limiter: limiter,
		logger:  logger,
	}
}

func (h *Handler) postRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientKey := GetClientKey(r)

	allowed, remaining, resetAt, err := h.limiter.Allow(ctx, clientKey)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := Response{
		Key:       clientKey,
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	if !allowed {
		metrics.RateLimitRequests.WithLabelValues("deny").Inc()
		w.WriteHeader(http.StatusTooManyRequests)
	}

	metrics.RateLimitRequests.WithLabelValues("allow").Inc()

	writeJSON(w, h.logger, resp)
}

func writeJSON(w http.ResponseWriter, logger *logi.Logger, resp Response) {
	body, err := json.Marshal(resp)
	if err != nil {
		logger.Error("json marshalling error", err)
		return
	}

	logger.LogDecision(resp.Key, resp.Remaining, resp.Allowed, resp.ResetAt)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		logger.Error("while trying to write response", err)
	}
}

func GetClientKey(r *http.Request) string {
	if auth := strings.TrimSpace(r.Header.Get("Authorization")); auth != "" {
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[len("bearer "):])
		}
		return auth
	}

	if apiKey := strings.TrimSpace(r.Header.Get("X-API-Key")); apiKey != "" {
		return apiKey
	}

	if cookie, err := r.Cookie("session_id"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return RealClientIP(r)
}

func RealClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
