package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrepare(t *testing.T) {
	dir := t.TempDir()
	args, err := prepare("python", "print(1)", dir)
	require.NoError(t, err)
	require.Equal(t, "docker", args[0])
	data, err := os.ReadFile(filepath.Join(dir, "solution.py"))
	require.NoError(t, err)
	require.Equal(t, "print(1)", string(data))

	dir = t.TempDir()
	args, err = prepare("go", "package main", dir)
	require.NoError(t, err)
	require.Contains(t, strings.Join(args, " "), "go run main.go")
	data, err = os.ReadFile(filepath.Join(dir, "main.go"))
	require.NoError(t, err)
	require.Equal(t, "package main", string(data))

	_, err = prepare("ruby", "puts 1", dir)
	require.Error(t, err)
}

func TestRunAndTruncate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := Run(ctx, RunRequest{Language: "ruby", Code: "puts 1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported language")

	result, err := Run(context.Background(), RunRequest{Language: "python", Code: "print(1)"})
	if err != nil {
		require.Contains(t, err.Error(), "docker run")
	} else {
		require.False(t, result.TimedOut)
	}

	require.Equal(t, "abc", truncate("abcdef", 3))
	require.Equal(t, "abc", truncate("abc", 3))
}
