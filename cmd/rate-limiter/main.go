package main

import (
	"os"
	"rate-limiter-go/internal/app"
	"rate-limiter-go/internal/config"
	"rate-limiter-go/internal/http"
	"rate-limiter-go/internal/log"
)

func main() {
	logger := log.NewLogger()
	conf, err := config.GetConfig()
	if err != nil {
		logger.Error("could not load config", err)
		os.Exit(1)
	}

	if err := conf.Validate(); err != nil {
		logger.Error("invalid configurations", err)
		os.Exit(1)
	}

	server := http.New(logger, conf.HTTPAddr)

	app := app.NewApp(server, conf, logger)
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
