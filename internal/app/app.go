package app

import (
	"fmt"
	"rate-limiter-go/internal/config"

	httpi "rate-limiter-go/internal/http"
	rlimiter "rate-limiter-go/internal/limiter/redis"
	logi "rate-limiter-go/internal/log"

	"github.com/redis/go-redis/v9"
)

type App struct {
	logger *logi.Logger
	server *httpi.Server
	config *config.Config
}

func NewApp(server *httpi.Server, config *config.Config, logger *logi.Logger) *App {
	return &App{
		logger: logger,
		server: server,
		config: config,
	}
}

func (a *App) Run() error {
	a.logger.Log(fmt.Sprintf("config: %v", a.config))

	conf := a.config
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddress,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})

	limiter := rlimiter.NewTokenBucketLimiter(
		redisClient,
		a.config.RateLimit,
		a.config.WindowSec,
		"rl",
	)

	logger := logi.NewLogger()
	logger.Init()

	httpi.RegisterEndpoints(logger, limiter, a.server.Router())

	logger.Log("Starting server")
	if err := a.server.Init(); err != nil {
		return err
	}

	return nil
}
