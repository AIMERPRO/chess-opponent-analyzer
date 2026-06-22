package main

import (
	"context"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/app"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/logger"
	"go.uber.org/zap"
)

// @title           Chess Opponent Analyzer API
// @version         1.0
// @description     API for analyzing chess opponents on lichess.org
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, _ := zap.NewDevelopment()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	log, err = logger.NewLogger(cfg.AppEnv)
	if err != nil {
		log.Fatal("failed to create logger", zap.Error(err))
	}

	application, err := app.NewApp(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed to create app", zap.Error(err))
	}

	if err = application.Start(); err != nil {
		log.Fatal("failed to start app", zap.Error(err))
	}
}
