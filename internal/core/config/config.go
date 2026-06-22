package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PgUser     string
	PgPass     string
	PgHost     string
	PgPort     int
	PgDatabase string
	PgSSLMode  string

	RedisHost string
	RedisPort int

	JwtSecret string

	GoPort             int
	AppEnv             string
	CORSAllowedOrigins string
	GlobalRateLimit    int
	GlobalRateBurst    int
	IPRateLimit        int
	IPRateBurst        int

	LichessGetGamesURL string
}

// Config Constructor from .env file

func NewConfig() (*Config, error) {
	/* Variables for POSTGRES */
	pgUser := os.Getenv("PG_USER")
	if pgUser == "" {
		return nil, fmt.Errorf("PG_USER environment variable not set")
	}
	pgPass := os.Getenv("PG_PASSWORD")
	if pgPass == "" {
		return nil, fmt.Errorf("PG_PASSWORD environment variable not set")
	}
	pgPort, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		return nil, fmt.Errorf("PG_PORT must be a valid integer: %w", err)
	}
	pgHost := os.Getenv("PG_HOST")
	if pgHost == "" {
		return nil, fmt.Errorf("PG_HOST environment variable not set")
	}
	pgDatabase := os.Getenv("PG_DATABASE")
	if pgDatabase == "" {
		return nil, fmt.Errorf("PG_DATABASE environment variable not set")
	}
	pgSSLMode := os.Getenv("PG_SSL_MODE")
	if pgSSLMode == "" {
		return nil, fmt.Errorf("PG_SSL_MODE environment variable not set")
	}

	/* Variables for REDIS */
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return nil, fmt.Errorf("REDIS_HOST environment variable not set")
	}
	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_PORT must be a valid integer: %w", err)
	}

	/* Variables for JWT */
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}

	/* Variables for GO Backend */
	goPort, err := strconv.Atoi(os.Getenv("GO_PORT"))
	if err != nil {
		return nil, fmt.Errorf("GO_PORT must be a valid integer: %w", err)
	}
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return nil, fmt.Errorf("APP_ENV environment variable not set")
	}

	corsAllowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsAllowedOrigins == "" {
		corsAllowedOrigins = "*"
	}

	globalRateLimit, err := strconv.Atoi(os.Getenv("GLOBAL_RATE_LIMIT"))
	if err != nil {
		return nil, fmt.Errorf("GLOBAL_RATE_LIMIT must be a valid integer: %w", err)
	}
	globalRateBurst, err := strconv.Atoi(os.Getenv("GLOBAL_RATE_BURST"))
	if err != nil {
		return nil, fmt.Errorf("GLOBAL_RATE_BURST must be a valid integer: %w", err)
	}

	ipRateLimit, err := strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	if err != nil {
		return nil, fmt.Errorf("IP_RATE_LIMIT must be a valid integer: %w", err)
	}
	ipRateBurst, err := strconv.Atoi(os.Getenv("IP_RATE_BURST"))
	if err != nil {
		return nil, fmt.Errorf("IP_RATE_BURST must be a valid integer: %w", err)
	}

	/* Variables for Lichess API */
	lichessGetGamesURL := os.Getenv("LICHESS_GET_GAMES_URL")
	if lichessGetGamesURL == "" {
		return nil, fmt.Errorf("LICHESS_GET_GAMES_URL environment variable not set")
	}

	return &Config{
		PgUser:     pgUser,
		PgPass:     pgPass,
		PgHost:     pgHost,
		PgPort:     pgPort,
		PgDatabase: pgDatabase,
		PgSSLMode:  pgSSLMode,

		RedisHost: redisHost,
		RedisPort: redisPort,

		JwtSecret: jwtSecret,

		GoPort:             goPort,
		AppEnv:             appEnv,
		CORSAllowedOrigins: corsAllowedOrigins,
		GlobalRateLimit:    globalRateLimit,
		GlobalRateBurst:    globalRateBurst,
		IPRateLimit:        ipRateLimit,
		IPRateBurst:        ipRateBurst,

		LichessGetGamesURL: lichessGetGamesURL,
	}, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PgUser, c.PgPass, c.PgHost, c.PgPort, c.PgDatabase, c.PgSSLMode,
	)
}
