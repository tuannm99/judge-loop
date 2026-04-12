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

	dir = t.TempDir()
	args, err = prepare("javascript", "console.log(1)", dir)
	require.NoError(t, err)
	require.Contains(t, strings.Join(args, " "), "node /sandbox/solution.js")
	data, err = os.ReadFile(filepath.Join(dir, "solution.js"))
	require.NoError(t, err)
	require.Equal(t, "console.log(1)", string(data))

	dir = t.TempDir()
	args, err = prepare("typescript", "console.log(1)", dir)
	require.NoError(t, err)
	require.Contains(t, strings.Join(args, " "), "deno run --quiet --no-check /sandbox/solution.ts")
	data, err = os.ReadFile(filepath.Join(dir, "solution.ts"))
	require.NoError(t, err)
	require.Equal(t, "console.log(1)", string(data))

	dir = t.TempDir()
	args, err = prepare("rust", "fn main() {}", dir)
	require.NoError(t, err)
	require.Contains(t, strings.Join(args, " "), "rustc -O main.rs -o /tmp/solution && /tmp/solution")
	data, err = os.ReadFile(filepath.Join(dir, "main.rs"))
	require.NoError(t, err)
	require.Equal(t, "fn main() {}", string(data))

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

func writeFakeDocker(t *testing.T, body string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o755))
	return dir
}

func TestRunReturnsExitCodeAndTruncatedStderr(t *testing.T) {
	fakeBin := writeFakeDocker(
		t,
		"#!/bin/sh\nprintf 'out'\npython3 - <<'PY'\nimport sys\nsys.stderr.write('x' * 700)\nPY\nexit 7\n",
	)
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	result, err := Run(context.Background(), RunRequest{Language: "python", Code: "print(1)"})
	require.NoError(t, err)
	require.Equal(t, "out", result.Output)
	require.Equal(t, 7, result.ExitCode)
	require.Len(t, result.Stderr, 500)
	require.False(t, result.TimedOut)
}

func TestRunMarksDeadlineExceededAsTimeout(t *testing.T) {
	fakeBin := writeFakeDocker(t, "#!/bin/sh\nsleep 1\n")
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result, err := Run(ctx, RunRequest{Language: "go", Code: "package main\nfunc main(){}"})
	require.NoError(t, err)
	require.True(t, result.TimedOut)
}

func TestRunReturnsDockerInvocationError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := Run(context.Background(), RunRequest{Language: "python", Code: "print(1)"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "docker run")
}

func TestPrepareWriteErrors(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing")

	_, err := prepare("python", "print(1)", missingDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write solution")

	_, err = prepare("go", "package main", missingDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write solution")

	_, err = prepare("javascript", "console.log(1)", missingDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write solution")

	_, err = prepare("typescript", "console.log(1)", missingDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write solution")

	_, err = prepare("rust", "fn main() {}", missingDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write solution")
}
