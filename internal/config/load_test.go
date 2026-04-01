package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetEnvLoader() {
	loadDotEnvOnce = sync.Once{}
}

func TestLoadAPIServerFromDotEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("PORT", "")
	t.Setenv("USER_ID", "")

	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".env"), []byte(
		"DATABASE_URL=postgres://from-dotenv\nREDIS_URL=redis-from-dotenv\nPORT=9090\nUSER_ID=11111111-1111-1111-1111-111111111111\n",
	), 0o644))
	t.Chdir(tmp)

	resetEnvLoader()
	cfg, err := LoadAPIServer()
	require.NoError(t, err)
	require.Equal(t, "postgres://from-dotenv", cfg.DatabaseURL)
	require.Equal(t, "redis-from-dotenv", cfg.RedisURL)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", cfg.UserID)
}

func TestLoadJudgeWorkerFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://worker")
	t.Setenv("REDIS_URL", "redis://worker")
	t.Setenv("CONCURRENCY", "7")
	t.Setenv("TIME_LIMIT_SECS", "15")

	resetEnvLoader()
	cfg, err := LoadJudgeWorker()
	require.NoError(t, err)
	require.Equal(t, "postgres://worker", cfg.DatabaseURL)
	require.Equal(t, "redis://worker", cfg.RedisURL)
	require.Equal(t, 7, cfg.Concurrency)
	require.Equal(t, 15, cfg.TimeLimitSecs)
}

func TestLoadLocalAgentDefaultsDataDir(t *testing.T) {
	t.Setenv("JUDGE_SERVER_URL", "http://localhost:9000")
	t.Setenv("JUDGE_PORT", "6060")
	t.Setenv("JUDGE_USER_ID", "22222222-2222-2222-2222-222222222222")
	t.Setenv("JUDGE_DATA_DIR", "")
	t.Setenv("JUDGE_REGISTRY_PATH", "/tmp/registry")

	resetEnvLoader()
	cfg, err := LoadLocalAgent()
	require.NoError(t, err)
	require.Equal(t, "http://localhost:9000", cfg.ServerURL)
	require.Equal(t, 6060, cfg.Port)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", cfg.UserID)
	require.Equal(t, "/tmp/registry", cfg.RegistryPath)
	require.Equal(t, defaultDataDir(), cfg.DataDir)
}

func TestLoadLocalAgentKeepsExplicitDataDir(t *testing.T) {
	t.Setenv("JUDGE_DATA_DIR", "/var/tmp/judge-loop")

	resetEnvLoader()
	cfg, err := LoadLocalAgent()
	require.NoError(t, err)
	require.Equal(t, "/var/tmp/judge-loop", cfg.DataDir)
}
