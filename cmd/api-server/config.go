package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// Config holds all runtime configuration for the api-server.
type Config struct {
	DatabaseURL string
	Port        string
	// UserID is the single hardcoded user for the no-auth MVP.
	UserID uuid.UUID
}

// LoadConfig reads configuration from environment variables with sane defaults
// for local development.
func LoadConfig() (Config, error) {
	dsn := env("DATABASE_URL", "postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable")
	port := env("PORT", "8080")
	rawUID := env("USER_ID", "00000000-0000-0000-0000-000000000001")

	uid, err := uuid.Parse(rawUID)
	if err != nil {
		return Config{}, fmt.Errorf("invalid USER_ID %q: %w", rawUID, err)
	}

	return Config{
		DatabaseURL: dsn,
		Port:        port,
		UserID:      uid,
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
