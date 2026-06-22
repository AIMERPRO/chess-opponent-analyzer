package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/features/analysis"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/features/auth"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/infrastructure/lichess"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/infrastructure/postgres"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/infrastructure/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type App struct {
	log         *zap.Logger
	cfg         *config.Config
	pool        *pgxpool.Pool
	redisClient *goredis.Client
	server      *http.Server
	router      *http.ServeMux
}

func NewApp(ctx context.Context, cfg *config.Config, log *zap.Logger) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	redisClient, err := redis.NewRedisClient(ctx, cfg.RedisHost, cfg.RedisPort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.GoPort),
	}

	router := http.NewServeMux()
	server.Handler = router

	authTokenRepo := auth.NewTokenRepo(pool)
	authUserRepo := auth.NewUserRepo(pool)
	authService := auth.NewService(authUserRepo, authTokenRepo, cfg)
	authRouter := auth.NewHandler(authService, log)
	authRouter.RegisterRoutes(router, cfg)

	lichessClient := lichess.NewClient(cfg.LichessGetGamesURL)
	analysisService := analysis.NewService(lichessClient, redisClient)
	analysisRouter := analysis.NewHandler(analysisService, log)
	analysisRouter.RegisterRoutes(router, cfg)

	return &App{
		log:         log,
		cfg:         cfg,
		pool:        pool,
		redisClient: redisClient,
		server:      server,
		router:      router,
	}, nil
}

func (a *App) Start() error {
	a.log.Info("Starting App")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		a.Stop()
	}()

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown server", zap.Error(err))
	}

	if err := a.redisClient.Close(); err != nil {
		a.log.Error("failed to close redis connection", zap.Error(err))
	}

	a.pool.Close()
}
