// Package sandbox runs user code in an isolated Docker container.
// It mounts a temp directory with the code file as read-only, then runs the
// appropriate interpreter/compiler inside the container with no network access.
package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunRequest describes a single code execution.
type RunRequest struct {
	Language string // "python" or "go"
	Code     string // source code
	Input    string // stdin fed to the program
}

// RunResult captures the outcome of one execution.
type RunResult struct {
	Output    string // stdout
	Stderr    string // raw stderr (first 500 chars)
	ExitCode  int
	TimedOut  bool
	RuntimeMS int64
}

// Run executes code in a Docker sandbox and returns the result.
// The caller should set a deadline on ctx for the TLE limit.
func Run(ctx context.Context, req RunRequest) (RunResult, error) {
	dir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return RunResult{}, fmt.Errorf("mktemp: %w", err)
	}
	defer os.RemoveAll(dir)

	dockerArgs, err := prepare(req.Language, req.Code, dir)
	if err != nil {
		return RunResult{}, err
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, dockerArgs[0], dockerArgs[1:]...) //nolint:gosec
	cmd.Stdin = strings.NewReader(req.Input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	runtimeMS := time.Since(start).Milliseconds()

	result := RunResult{
		Output:    stdout.String(),
		Stderr:    truncate(stderr.String(), 500),
		RuntimeMS: runtimeMS,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result, nil
	}

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("docker run: %w", runErr)
		}
	}

	return result, nil
}

// prepare writes the source file to dir and returns the docker run command args.
func prepare(language, code, dir string) ([]string, error) {
	base := []string{
		"docker", "run", "--rm", "-i",
		"--network", "none",
		"--pids-limit", "50",
		"-v", dir + ":/sandbox:ro",
	}

	switch strings.ToLower(language) {
	case "python":
		if err := os.WriteFile(filepath.Join(dir, "solution.py"), []byte(code), 0o644); err != nil {
			return nil, fmt.Errorf("write solution: %w", err)
		}
		return append(base,
			"--memory", "64m", "--memory-swap", "64m",
			"--cpus", "0.5",
			"python:3.12-slim",
			"python3", "/sandbox/solution.py",
		), nil

	case "go":
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0o644); err != nil {
			return nil, fmt.Errorf("write solution: %w", err)
		}
		return append(base,
			"--memory", "256m", "--memory-swap", "256m",
			"--cpus", "0.5",
			"golang:1.25-alpine",
			"sh", "-c", "cd /sandbox && go run main.go",
		), nil

	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

func truncate(s string, m int) string {
	if len(s) <= m {
		return s
	}
	return s[:m]
}
