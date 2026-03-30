package main

import "os"

// Config holds runtime configuration for the judge-worker.
type Config struct {
	DatabaseURL   string
	RedisURL      string
	Concurrency   int
	TimeLimitSecs int // per test-case execution timeout
}

// LoadConfig reads config from environment variables with local-dev defaults.
func LoadConfig() Config {
	return Config{
		DatabaseURL:   envOr("DATABASE_URL", "postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable"),
		RedisURL:      envOr("REDIS_URL", "localhost:6379"),
		Concurrency:   envInt("CONCURRENCY", 2),
		TimeLimitSecs: envInt("TIME_LIMIT_SECS", 10),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return fallback
		}
		n = n*10 + int(c-'0')
	}
	return n
}
