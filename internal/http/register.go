package http

import (
	"net/http"
	"rate-limiter-go/internal/limiter"
	logi "rate-limiter-go/internal/log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	metrics "rate-limiter-go/internal/metrics"

	"github.com/labstack/echo/v4"
)

func RegisterEndpoints(logger *logi.Logger, limiter limiter.Limiter, router *echo.Echo) {
	h := newHandler(limiter, logger)
	router.POST("/v1/rate-limit", func(c echo.Context) error {
		h.postRateLimit(c.Response(), c.Request())
		return nil
	})

	router.GET("/healthz", func(c echo.Context) error {
		logger.Log("everything is ok")
		return c.String(http.StatusOK, "Health's ok")
	})

	router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			status := c.Response().Status
			metrics.HTTPReqDuration.
				WithLabelValues(c.Path(), c.Request().Method, strconv.Itoa(status)).
				Observe(time.Since(start).Seconds())
			return err
		}
	})
}
