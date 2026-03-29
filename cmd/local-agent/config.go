package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration for the local-agent.
type Config struct {
	ServerURL string `yaml:"server_url"`
	Port      int    `yaml:"port"`
	UserID    string `yaml:"user_id"`
	DataDir   string `yaml:"data_dir"`
}

// LoadConfig builds a Config by:
//  1. Starting with defaults suitable for local development.
//  2. Merging in values from ~/.config/judge-loop/agent.yaml (warning if missing).
//  3. Applying env var overrides (JUDGE_SERVER_URL, JUDGE_PORT, JUDGE_USER_ID).
func LoadConfig() Config {
	cfg := Config{
		ServerURL: "http://localhost:8080",
		Port:      7070,
		UserID:    "00000000-0000-0000-0000-000000000001",
		DataDir:   defaultDataDir(),
	}

	if path := configFilePath(); path != "" {
		if err := loadYAML(path, &cfg); err != nil {
			log.Printf("config: warning: %v (using defaults)", err)
		}
	}

	applyEnv(&cfg)
	return cfg
}

func loadYAML(path string, cfg *Config) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found at %s", path)
	}
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(cfg)
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("JUDGE_SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("JUDGE_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Port = n
		}
	}
	if v := os.Getenv("JUDGE_USER_ID"); v != "" {
		cfg.UserID = v
	}
}

func configFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "judge-loop", "agent.yaml")
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".judge-loop"
	}
	return filepath.Join(home, ".local", "share", "judge-loop")
}
