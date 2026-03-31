package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var loadDotEnvOnce sync.Once

func parse(target any) error {
	loadDotEnvOnce.Do(loadDotEnv)
	return env.Parse(target)
}

func loadDotEnv() {
	_ = godotenv.Load()

	if cwd, err := os.Getwd(); err == nil {
		_ = godotenv.Overload(filepath.Join(cwd, ".env"))
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".judge-loop"
	}
	return filepath.Join(home, ".local", "share", "judge-loop")
}
